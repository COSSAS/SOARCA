package queue

import (
	"context"
	"sync"
	"testing"
	"time"

	"soarca/pkg/models/fin"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func newJob(capabilityType string) fin.Job {
	return fin.Job{
		JobId:                 uuid.New(),
		ExecutionId:           uuid.New(),
		StepId:                "step--1",
		StepExecutionId:       uuid.New(),
		CapabilityType:        capabilityType,
		LeaseExpiresInSeconds: 60,
	}
}

func TestEnqueueClaimSubmitRoundTrip(t *testing.T) {
	q := New()
	defer q.Close()

	job := newJob("pong")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	var result fin.JobResult
	var enqueueErr error
	go func() {
		defer wg.Done()
		result, enqueueErr = q.Enqueue(ctx, job)
	}()

	claimed, err := q.Claim(ctx, []string{"pong"}, "fin-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, claimed.JobId, job.JobId)

	err = q.Submit(job.JobId, "fin-1", fin.JobResult{State: fin.JobStateSuccess})
	if err != nil {
		t.Fatal(err)
	}

	wg.Wait()
	if enqueueErr != nil {
		t.Fatal(enqueueErr)
	}
	assert.Equal(t, result.State, fin.JobStateSuccess)
}

func TestClaimBlocksUntilJobIsEnqueued(t *testing.T) {
	q := New()
	defer q.Close()

	claimCtx, claimCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer claimCancel()

	claimed := make(chan fin.Job, 1)
	go func() {
		job, err := q.Claim(claimCtx, []string{"pong"}, "fin-1")
		if err == nil {
			claimed <- job
		}
	}()

	// Give the Claim goroutine time to start blocking before anything is
	// enqueued, to actually exercise the long-poll wait path.
	time.Sleep(50 * time.Millisecond)

	job := newJob("pong")
	enqueueCtx, enqueueCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer enqueueCancel()
	go func() { _, _ = q.Enqueue(enqueueCtx, job) }()

	select {
	case got := <-claimed:
		assert.Equal(t, got.JobId, job.JobId)
	case <-time.After(2 * time.Second):
		t.Fatal("Claim never returned the job enqueued while it was blocked")
	}
}

func TestClaimOnlyMatchesRequestedCapabilityTypes(t *testing.T) {
	q := New()
	defer q.Close()

	pingJob := newJob("ping")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() { _, _ = q.Enqueue(ctx, pingJob) }()

	time.Sleep(20 * time.Millisecond)

	claimCtx, claimCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer claimCancel()
	_, err := q.Claim(claimCtx, []string{"pong"}, "fin-1")
	if err == nil {
		t.Fatal("expected Claim to time out: no job of a matching capability type is queued")
	}
}

func TestEnqueueContextDoneRemovesUnclaimedJob(t *testing.T) {
	q := New()
	defer q.Close()

	job := newJob("pong")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := q.Enqueue(ctx, job)
	if err == nil {
		t.Fatal("expected Enqueue to return a context error once its deadline passes with nobody claiming the job")
	}

	// The job must have been removed: a subsequent claim attempt with a
	// short-lived context must time out rather than finding it.
	claimCtx, claimCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer claimCancel()
	_, err = q.Claim(claimCtx, []string{"pong"}, "fin-1")
	if err == nil {
		t.Fatal("expected no job to be claimable after its Enqueue context expired")
	}
}

func TestEnqueueContextDoneRemovesLeasedJob(t *testing.T) {
	q := New()
	defer q.Close()

	job := newJob("pong")
	enqueueCtx, enqueueCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer enqueueCancel()

	done := make(chan struct{})
	go func() {
		_, _ = q.Enqueue(enqueueCtx, job)
		close(done)
	}()

	claimCtx, claimCancel := context.WithTimeout(context.Background(), time.Second)
	defer claimCancel()
	_, err := q.Claim(claimCtx, []string{"pong"}, "fin-1")
	if err != nil {
		t.Fatal(err)
	}

	<-done // wait for the Enqueue caller's deadline to pass

	// The Fin that claimed the job before it was abandoned should no longer
	// be able to submit a result for it.
	err = q.Submit(job.JobId, "fin-1", fin.JobResult{State: fin.JobStateSuccess})
	if err == nil {
		t.Fatal("expected Submit to fail: the job was removed once its Enqueue context expired")
	}
}

func TestSubmitFailsForUnknownJob(t *testing.T) {
	q := New()
	defer q.Close()

	err := q.Submit(uuid.New(), "fin-1", fin.JobResult{State: fin.JobStateSuccess})
	if err == nil {
		t.Fatal("expected Submit to fail for a job id that was never enqueued")
	}
}

func TestSubmitFailsWhenLeasedToADifferentFin(t *testing.T) {
	q := New()
	defer q.Close()

	job := newJob("pong")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() { _, _ = q.Enqueue(ctx, job) }()

	_, err := q.Claim(ctx, []string{"pong"}, "fin-1")
	if err != nil {
		t.Fatal(err)
	}

	err = q.Submit(job.JobId, "fin-2", fin.JobResult{State: fin.JobStateSuccess})
	if err == nil {
		t.Fatal("expected Submit to fail: job is leased to fin-1, not fin-2")
	}
}

func TestExpiredLeaseIsRequeuedForAnotherFin(t *testing.T) {
	q := New()
	defer q.Close()

	job := newJob("pong")
	job.LeaseExpiresInSeconds = 0 // forces the minimum, still short, lease via a tiny override below
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() { _, _ = q.Enqueue(ctx, job) }()

	claimed, err := q.Claim(ctx, []string{"pong"}, "fin-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, claimed.JobId, job.JobId)

	// Force the lease to already be expired, instead of waiting out the
	// real default lease duration, then let the sweeper's next tick pick it
	// up.
	q.mu.Lock()
	if e, ok := q.leased[job.JobId]; ok {
		e.leaseExpiry = time.Now().Add(-time.Second)
	}
	q.mu.Unlock()

	claimCtx, claimCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer claimCancel()
	reclaimed, err := q.Claim(claimCtx, []string{"pong"}, "fin-2")
	if err != nil {
		t.Fatal("expected the expired lease to be requeued and claimable by another fin:", err)
	}
	assert.Equal(t, reclaimed.JobId, job.JobId)

	err = q.Submit(job.JobId, "fin-2", fin.JobResult{State: fin.JobStateSuccess})
	if err != nil {
		t.Fatal(err)
	}
}

func TestExtendLeasePreventsExpiryRequeue(t *testing.T) {
	q := New()
	defer q.Close()

	job := newJob("pong")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go func() { _, _ = q.Enqueue(ctx, job) }()

	_, err := q.Claim(ctx, []string{"pong"}, "fin-1")
	if err != nil {
		t.Fatal(err)
	}

	// Simulate a lease about to expire, then extend it.
	q.mu.Lock()
	q.leased[job.JobId].leaseExpiry = time.Now().Add(10 * time.Millisecond)
	q.mu.Unlock()

	if err := q.ExtendLease(job.JobId, "fin-1", 5); err != nil {
		t.Fatal(err)
	}

	// Give the sweeper a couple of ticks to run; the job must still be
	// leased to fin-1, not requeued.
	time.Sleep(2*defaultSweepInterval + 100*time.Millisecond)

	err = q.Submit(job.JobId, "fin-1", fin.JobResult{State: fin.JobStateSuccess})
	if err != nil {
		t.Fatal("expected job to still be leased to fin-1 after ExtendLease:", err)
	}
}

func TestExtendLeaseFailsForUnknownJobOrWrongFin(t *testing.T) {
	q := New()
	defer q.Close()

	if err := q.ExtendLease(uuid.New(), "fin-1", 30); err == nil {
		t.Fatal("expected ExtendLease to fail for an unknown job id")
	}

	job := newJob("pong")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() { _, _ = q.Enqueue(ctx, job) }()

	_, err := q.Claim(ctx, []string{"pong"}, "fin-1")
	if err != nil {
		t.Fatal(err)
	}

	if err := q.ExtendLease(job.JobId, "fin-2", 30); err == nil {
		t.Fatal("expected ExtendLease to fail: job is leased to fin-1, not fin-2")
	}
}
