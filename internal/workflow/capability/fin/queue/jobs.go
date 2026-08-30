// Package queue implements the in-memory Fin job queue.
//
// It is intentionally not persisted because in-flight runs are not
// resumed across restarts either.
package queue

import (
	"context"
	"reflect"
	"sync"
	"time"

	"soarca/internal/logger"
	"soarca/pkg/fins/protocol"

	"github.com/google/uuid"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// defaultLeaseDuration is used when a Job does not specify a positive
// LeaseExpiresInSeconds.
const defaultLeaseDuration = 60 * time.Second

// defaultSweepInterval is how often expired leases are checked and
// requeued.
const defaultSweepInterval = time.Second

// entry is the queue's internal bookkeeping for one job: the job payload
// itself, the channel its Enqueue caller is blocked reading from, and its
// current lease state (unleased jobs have an empty LeasedTo).
type entry struct {
	job         fin.Job
	resultCh    chan fin.JobResult
	leasedTo    string
	leaseExpiry time.Time
}

// Queue is the in-memory Fin job queue: a set of per-CapabilityType FIFO
// queues of unleased jobs, plus a map of currently-leased jobs keyed by
// JobId. Safe for concurrent use.
type Queue struct {
	mu      sync.Mutex
	pending map[string][]*entry  // capability type -> FIFO of unleased jobs
	leased  map[uuid.UUID]*entry // job id -> currently-leased job
	notify  chan struct{}        // closed and replaced whenever queue state changes, to wake blocked Claim callers
	stop    chan struct{}
	done    chan struct{}
}

// New creates an empty Queue and starts its background lease-expiry
// sweeper. Call Close to stop the sweeper goroutine (e.g. on shutdown, or
// in tests).
func New() *Queue {
	q := &Queue{
		pending: map[string][]*entry{},
		leased:  map[uuid.UUID]*entry{},
		notify:  make(chan struct{}),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
	go q.sweepLoop()
	return q
}

// Close stops the background lease-expiry sweeper. It does not affect
// already-enqueued or leased jobs.
func (q *Queue) Close() {
	close(q.stop)
	<-q.done
}

// Enqueue registers job under its CapabilityType and blocks until a
// claiming Fin submits a result, or ctx is done - whichever comes first. If
// ctx is done first, job is removed from the queue (wherever it currently
// is: still pending, or leased but not yet resolved) so a late claim/result
// can no longer observe or resolve it.
func (q *Queue) Enqueue(ctx context.Context, job fin.Job) (fin.JobResult, error) {
	e := &entry{job: job, resultCh: make(chan fin.JobResult, 1)}

	q.mu.Lock()
	q.pending[job.CapabilityType] = append(q.pending[job.CapabilityType], e)
	q.broadcastLocked()
	q.mu.Unlock()

	log.Trace("enqueued job ", job.JobId.String(), " for capability type ", job.CapabilityType)

	select {
	case result := <-e.resultCh:
		return result, nil
	case <-ctx.Done():
		q.removeJob(job.JobId, job.CapabilityType)
		return fin.JobResult{}, ctx.Err()
	}
}

// Claim blocks (long-poll) until a job is available under one of
// capabilityTypes, or ctx is done - whichever comes first. On success, the
// returned job is leased to finId until its lease expires or a result is
// submitted for it.
func (q *Queue) Claim(ctx context.Context, capabilityTypes []string, finId string) (fin.Job, error) {
	for {
		q.mu.Lock()
		for _, capabilityType := range capabilityTypes {
			jobs := q.pending[capabilityType]
			if len(jobs) == 0 {
				continue
			}
			e := jobs[0]
			q.pending[capabilityType] = jobs[1:]
			if len(q.pending[capabilityType]) == 0 {
				delete(q.pending, capabilityType)
			}
			e.leasedTo = finId
			e.leaseExpiry = time.Now().Add(leaseDuration(e.job))
			q.leased[e.job.JobId] = e
			q.mu.Unlock()
			log.Trace("claimed job ", e.job.JobId.String(), " for fin ", finId)
			return e.job, nil
		}
		waitCh := q.notify
		q.mu.Unlock()

		select {
		case <-waitCh:
			continue
		case <-ctx.Done():
			return fin.Job{}, ctx.Err()
		}
	}
}

// Submit delivers result for jobId to the Enqueue caller waiting on it,
// provided jobId is currently leased to finId. If the Enqueue caller has
// already given up (ctx done), result is dropped silently - there is
// nothing left to deliver it to.
func (q *Queue) Submit(jobId uuid.UUID, finId string, result fin.JobResult) error {
	q.mu.Lock()
	e, ok := q.leased[jobId]
	if !ok {
		q.mu.Unlock()
		return fin.ErrJobNotFound{JobId: jobId.String()}
	}
	if e.leasedTo != finId {
		q.mu.Unlock()
		return fin.ErrJobNotLeasedToFin{JobId: jobId.String(), FinId: finId}
	}
	delete(q.leased, jobId)
	q.mu.Unlock()

	select {
	case e.resultCh <- result:
	default:
		log.Warning("result submitted for job ", jobId.String(), " but nothing is waiting for it anymore (deadline already exceeded); dropping")
	}
	return nil
}

// ExtendLease refreshes a leased job's timeout for status-ping-driven work.
func (q *Queue) ExtendLease(jobId uuid.UUID, finId string, extendBySeconds int) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	e, ok := q.leased[jobId]
	if !ok {
		return fin.ErrJobNotFound{JobId: jobId.String()}
	}
	if e.leasedTo != finId {
		return fin.ErrJobNotLeasedToFin{JobId: jobId.String(), FinId: finId}
	}
	duration := time.Duration(extendBySeconds) * time.Second
	if extendBySeconds <= 0 {
		duration = defaultLeaseDuration
	}
	e.leaseExpiry = time.Now().Add(duration)
	return nil
}

// removeJob deletes jobId from wherever it currently sits (pending or
// leased). Used when an Enqueue caller's ctx is done before a result
// arrives.
func (q *Queue) removeJob(jobId uuid.UUID, capabilityType string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if jobs, ok := q.pending[capabilityType]; ok {
		for i, e := range jobs {
			if e.job.JobId == jobId {
				q.pending[capabilityType] = append(jobs[:i], jobs[i+1:]...)
				if len(q.pending[capabilityType]) == 0 {
					delete(q.pending, capabilityType)
				}
				break
			}
		}
	}
	delete(q.leased, jobId)
}

// sweepLoop periodically requeues jobs whose lease has expired without a
// result or status ping, so another Fin registered under the same
// CapabilityType can claim them.
func (q *Queue) sweepLoop() {
	defer close(q.done)
	ticker := time.NewTicker(defaultSweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-q.stop:
			return
		case <-ticker.C:
			q.sweepExpiredLeases()
		}
	}
}

func (q *Queue) sweepExpiredLeases() {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	requeued := false
	for jobId, e := range q.leased {
		if now.Before(e.leaseExpiry) {
			continue
		}
		log.Warning("lease expired for job ", jobId.String(), " (was leased to fin ", e.leasedTo, "); requeuing for capability type ", e.job.CapabilityType)
		delete(q.leased, jobId)
		e.leasedTo = ""
		q.pending[e.job.CapabilityType] = append(q.pending[e.job.CapabilityType], e)
		requeued = true
	}
	if requeued {
		q.broadcastLocked()
	}
}

// broadcastLocked wakes every Claim call currently blocked waiting for new
// work. Must be called with q.mu held.
func (q *Queue) broadcastLocked() {
	close(q.notify)
	q.notify = make(chan struct{})
}

func leaseDuration(job fin.Job) time.Duration {
	if job.LeaseExpiresInSeconds <= 0 {
		return defaultLeaseDuration
	}
	return time.Duration(job.LeaseExpiresInSeconds) * time.Second
}
