package fin

import (
	"context"
	"time"

	"github.com/google/uuid"

	"soarca/internal/store"
	"soarca/internal/workflow/capability/fin/queue"
	"soarca/internal/workflow/capability/fin/token"
	"soarca/pkg/fins/protocol"
)

// WorkService implements the services.FinWorkService interface.
// It manages leased FIN work items, handling polling, result submission, and heartbeats.
type WorkService struct {
	store  storage.FinStore
	queue  *queue.Queue
	config WorkServiceConfig
}

// WorkServiceConfig holds configuration for the FIN work service.
type WorkServiceConfig struct {
	LongPollTimeoutSeconds int
	JobLeaseSeconds        int
}

// NewWorkService creates a new FinWorkService.
func NewWorkService(store storage.FinStore, q *queue.Queue, config WorkServiceConfig) *WorkService {
	return &WorkService{
		store:  store,
		queue:  q,
		config: config,
	}
}

// PollJob polls for available work matching the FIN's capabilities.
// Updates the FIN's last-seen timestamp.
func (w *WorkService) PollJob(ctx context.Context, finToken string, pollReq fin.PollRequest) (*fin.Job, error) {
	record, err := w.store.GetByTokenHash(ctx, token.Hash(finToken))
	if err != nil {
		return nil, err
	}

	// Update last-seen timestamp (touch the FIN).
	if err := w.store.Touch(ctx, record.FinId, time.Now()); err != nil {
		log.Warning("failed to update last-seen for fin ", record.FinId, ": ", err)
	}

	// Calculate poll timeout.
	timeout := time.Duration(w.config.LongPollTimeoutSeconds) * time.Second
	if w.config.LongPollTimeoutSeconds <= 0 {
		timeout = 25 * time.Second
	}
	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Claim a job matching the FIN's capabilities.
	job, err := w.queue.Claim(pollCtx, capabilityTypes(record.Capabilities), record.FinId)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// SubmitJobResult submits the result of a claimed job.
// Updates the FIN's last-seen timestamp.
func (w *WorkService) SubmitJobResult(ctx context.Context, finToken string, jobID uuid.UUID, result fin.JobResult) error {
	record, err := w.store.GetByTokenHash(ctx, token.Hash(finToken))
	if err != nil {
		return err
	}

	// Submit the result to the queue.
	if err := w.queue.Submit(jobID, record.FinId, result); err != nil {
		return err
	}

	// Update last-seen timestamp.
	if err := w.store.Touch(ctx, record.FinId, time.Now()); err != nil {
		log.Warning("failed to update last-seen for fin ", record.FinId, ": ", err)
	}

	return nil
}

// HeartbeatJob extends the lease on an in-flight job (status-ping).
// Updates the FIN's last-seen timestamp.
func (w *WorkService) HeartbeatJob(ctx context.Context, finToken string, jobID uuid.UUID) error {
	record, err := w.store.GetByTokenHash(ctx, token.Hash(finToken))
	if err != nil {
		return err
	}

	// Extend the lease.
	if err := w.queue.ExtendLease(jobID, record.FinId, w.config.JobLeaseSeconds); err != nil {
		return err
	}

	// Update last-seen timestamp.
	if err := w.store.Touch(ctx, record.FinId, time.Now()); err != nil {
		log.Warning("failed to update last-seen for fin ", record.FinId, ": ", err)
	}

	return nil
}
