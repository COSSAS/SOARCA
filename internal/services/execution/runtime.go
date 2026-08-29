package execution

import (
	"context"

	"github.com/google/uuid"

	"soarca/internal/controller/decomposer_controller"
	"soarca/pkg/core/decomposer"
	"soarca/pkg/models/cacao"
	modelcache "soarca/pkg/models/cache"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"
)

// ManualResumer resolves a paused manual step.
type ManualResumer interface {
	PostContinue(response manual.InteractionResponse) error
}

// ExecutionReports reads recorded execution state.
type ExecutionReports interface {
	GetExecutionReport(executionID uuid.UUID) (modelcache.ExecutionEntry, error)
}

// Service implements the ExecutionRuntime contract.
type Service struct {
	decomposers decomposer_controller.IController
	manual      ManualResumer
	reports     ExecutionReports
}

// New creates a new execution runtime service.
func New(decomposers decomposer_controller.IController, manualResumer ManualResumer, reports ExecutionReports) *Service {
	return &Service{
		decomposers: decomposers,
		manual:      manualResumer,
		reports:     reports,
	}
}

// StartExecution launches a playbook execution and waits for the execution ID.
func (s *Service) StartExecution(ctx context.Context, playbook *cacao.Playbook, variables cacao.Variables) (executionID uuid.UUID, err error) {
	_ = variables
	decomp := s.decomposers.NewDecomposer()
	executions := make(chan decomposer.ExecutionDetails, 1)

	go decomp.ExecuteAsync(*playbook, executions)

	for {
		select {
		case <-ctx.Done():
			return uuid.Nil, ctx.Err()
		case details := <-executions:
			if details.PlaybookId != playbook.ID {
				continue
			}
			return details.ExecutionId, nil
		}
	}
}

// ResumeManualStep resumes a paused manual step.
func (s *Service) ResumeManualStep(ctx context.Context, execID uuid.UUID, stepExecID uuid.UUID, response manual.InteractionResponse) error {
	_ = ctx
	response.Metadata = execution.Metadata{
		ExecutionId:     execID,
		StepExecutionId: stepExecID,
	}
	return s.manual.PostContinue(response)
}

// GetExecutionStatus returns the cached execution report for an execution.
func (s *Service) GetExecutionStatus(ctx context.Context, execID uuid.UUID) (modelcache.ExecutionEntry, error) {
	_ = ctx
	return s.reports.GetExecutionReport(execID)
}
