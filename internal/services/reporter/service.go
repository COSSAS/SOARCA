package reporter

import (
	"context"

	"soarca/internal/controller/informer"
	"soarca/pkg/models/cache"

	"github.com/google/uuid"
)

// Service implements the services.ReporterService interface.
type Service struct {
	informer informer.IExecutionInformer
}

// New creates a reporter service.
func New(informer informer.IExecutionInformer) *Service {
	return &Service{informer: informer}
}

// ListExecutions returns all recorded execution summaries.
func (s *Service) ListExecutions(ctx context.Context) ([]cache.ExecutionEntry, error) {
	_ = ctx
	return s.informer.GetExecutions()
}

// GetExecutionReport retrieves a detailed execution report.
func (s *Service) GetExecutionReport(ctx context.Context, executionID uuid.UUID) (cache.ExecutionEntry, error) {
	_ = ctx
	return s.informer.GetExecutionReport(executionID)
}
