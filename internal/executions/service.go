// Package executions owns playbook execution: starting runs, and reading back
// their recorded state.
package executions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"soarca/internal/controller/decomposer_controller"
	"soarca/internal/storage"
	"soarca/pkg/core/decomposer"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/cache"
)

// Runner is the execution use case surface offered to any driver
// (HTTP, gRPC, CLI, embedded SDK).
type Runner interface {
	// Start executes a playbook supplied by the caller.
	Start(ctx context.Context, playbook *cacao.Playbook, variables cacao.Variables) (executionID uuid.UUID, err error)

	// StartByID executes a stored playbook.
	StartByID(ctx context.Context, playbookID string, variables cacao.Variables) (executionID uuid.UUID, err error)

	// List returns all recorded execution summaries.
	List(ctx context.Context) ([]cache.ExecutionEntry, error)

	// Report returns the detailed report for a single execution.
	Report(ctx context.Context, executionID uuid.UUID) (cache.ExecutionEntry, error)
}

// Reports reads recorded execution state.
type Reports interface {
	GetExecutions() ([]cache.ExecutionEntry, error)
	GetExecutionReport(executionID uuid.UUID) (cache.ExecutionEntry, error)
}

// ValidationError marks a rejected execution request.
type ValidationError struct {
	Err error
}

func (e ValidationError) Error() string {
	return e.Err.Error()
}

func (e ValidationError) Unwrap() error {
	return e.Err
}

// Service implements Runner.
type Service struct {
	engine    decomposer_controller.IController
	playbooks storage.PlaybookStore
	reports   Reports
}

func New(engine decomposer_controller.IController, playbooks storage.PlaybookStore, reports Reports) *Service {
	return &Service{
		engine:    engine,
		playbooks: playbooks,
		reports:   reports,
	}
}

// StartByID loads a stored playbook, applies the supplied variables and starts execution.
func (s *Service) StartByID(ctx context.Context, playbookID string, variables cacao.Variables) (uuid.UUID, error) {
	playbook, err := s.playbooks.Get(ctx, playbookID)
	if err != nil {
		return uuid.Nil, err
	}
	return s.Start(ctx, &playbook, variables)
}

// Start applies the supplied variables to the playbook and starts execution.
func (s *Service) Start(ctx context.Context, playbook *cacao.Playbook, variables cacao.Variables) (uuid.UUID, error) {
	if err := mergeVariablesInPlaybook(playbook, variables); err != nil {
		return uuid.Nil, ValidationError{Err: err}
	}

	decomp := s.engine.NewDecomposer()
	details := make(chan decomposer.ExecutionDetails, 1)

	go decomp.ExecuteAsync(*playbook, details)

	select {
	case <-ctx.Done():
		return uuid.Nil, ctx.Err()
	case d := <-details:
		return d.ExecutionId, nil
	}
}

// List returns all recorded execution summaries.
func (s *Service) List(ctx context.Context) ([]cache.ExecutionEntry, error) {
	_ = ctx
	return s.reports.GetExecutions()
}

// Report returns the detailed report for a single execution.
func (s *Service) Report(ctx context.Context, executionID uuid.UUID) (cache.ExecutionEntry, error) {
	_ = ctx
	return s.reports.GetExecutionReport(executionID)
}

func mergeVariablesInPlaybook(playbook *cacao.Playbook, payloadVariables cacao.Variables) error {
	for name, variable := range payloadVariables {
		if _, ok := playbook.PlaybookVariables[name]; !ok {
			return fmt.Errorf("provided variables is not a valid subset of the variables for the referenced playbook [ playbook id: %s ]", playbook.ID)
		}
		if variable.Type != playbook.PlaybookVariables[name].Type {
			return fmt.Errorf("mismatch in variables type for [ %s ]: payload var type = %s, playbook var type = %s", name, variable.Type, playbook.PlaybookVariables[name].Type)
		}
		if !playbook.PlaybookVariables[name].External {
			return fmt.Errorf("playbook variable [ %s ] cannot be assigned in playbook because it is not marked as external in the plabook", name)
		}

		updatedVariable := cacao.Variable{
			Name:        name,
			Type:        playbook.PlaybookVariables[name].Type,
			Description: playbook.PlaybookVariables[name].Description,
			Value:       variable.Value,
			Constant:    playbook.PlaybookVariables[name].Constant,
			External:    playbook.PlaybookVariables[name].External,
		}
		playbook.PlaybookVariables[name] = updatedVariable
	}
	return nil
}

// DecodeVariables decodes a JSON trigger payload into variables.
func DecodeVariables(body []byte) (cacao.Variables, error) {
	payloadVariables := cacao.NewVariables()
	if err := json.Unmarshal(body, &payloadVariables); err != nil {
		return nil, errors.New("cannot unmarshal provided variables")
	}
	return payloadVariables, nil
}

// DecodePlaybook parses a playbook payload.
func DecodePlaybook(body []byte) *cacao.Playbook {
	return cacao.Decode(body)
}
