package execution

import (
	"context"

	"github.com/google/uuid"

	"soarca/internal/controller/decomposer_controller"
	"soarca/internal/logger"
	appruntime "soarca/internal/runtime"
	"soarca/pkg/core/decomposer"
	"soarca/pkg/models/cacao"
	modelcache "soarca/pkg/models/cache"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"
)

var log *logger.Log

// Service implements the ExecutionRuntime contract.
type Service struct {
	runtime    *appruntime.Runtime
	controller decomposer_controller.IController
}

// New creates a new execution runtime service.
func New(runtime *appruntime.Runtime, controller decomposer_controller.IController) *Service {
	return &Service{
		runtime:    runtime,
		controller: controller,
	}
}

// StartExecution launches a playbook execution and waits for the execution ID.
func (s *Service) StartExecution(ctx context.Context, playbook *cacao.Playbook, variables cacao.Variables) (executionID uuid.UUID, err error) {
	_ = variables
	decomp := s.controller.NewDecomposer()
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
	return s.runtime.GetInteraction().PostContinue(response)
}

// GetExecutionStatus returns the cached execution report for an execution.
func (s *Service) GetExecutionStatus(ctx context.Context, execID uuid.UUID) (modelcache.ExecutionEntry, error) {
	_ = ctx
	return s.runtime.GetCache().GetExecutionReport(execID)
}
