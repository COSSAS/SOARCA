// Package runs owns playbook runs: starting them, and reading back their
// recorded state.
package runs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"soarca/internal/store"
	"soarca/internal/workflow"
	"soarca/pkg/cacao"
	"soarca/internal/runs/state"
)

// Runner is the run use case surface offered to any driver
// (HTTP, gRPC, CLI, embedded SDK).
type Runner interface {
	// Start runs a playbook supplied by the caller.
	Start(ctx context.Context, playbook *cacao.Playbook, variables cacao.Variables) (runID uuid.UUID, err error)

	// StartByID runs a stored playbook.
	StartByID(ctx context.Context, playbookID string, variables cacao.Variables) (runID uuid.UUID, err error)

	// List returns all recorded run summaries.
	List(ctx context.Context) ([]runstate.RunEntry, error)

	// Report returns the detailed report for a single run.
	Report(ctx context.Context, runID uuid.UUID) (runstate.RunEntry, error)
}

// Reports reads recorded run state.
type Reports interface {
	GetRuns() ([]runstate.RunEntry, error)
	GetRunReport(runID uuid.UUID) (runstate.RunEntry, error)
}

// ValidationError marks a rejected run request.
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
	newWalker workflow.NewWalker
	playbooks storage.PlaybookStore
	reports   Reports
}

func New(newWalker workflow.NewWalker, playbooks storage.PlaybookStore, reports Reports) *Service {
	return &Service{
		newWalker: newWalker,
		playbooks: playbooks,
		reports:   reports,
	}
}

// StartByID loads a stored playbook, applies the supplied variables and starts a run.
func (s *Service) StartByID(ctx context.Context, playbookID string, variables cacao.Variables) (uuid.UUID, error) {
	playbook, err := s.playbooks.Get(ctx, playbookID)
	if err != nil {
		return uuid.Nil, err
	}
	return s.Start(ctx, &playbook, variables)
}

// Start applies the supplied variables to the playbook and starts a run.
func (s *Service) Start(ctx context.Context, playbook *cacao.Playbook, variables cacao.Variables) (uuid.UUID, error) {
	if err := mergeVariablesInPlaybook(playbook, variables); err != nil {
		return uuid.Nil, ValidationError{Err: err}
	}

	walker := s.newWalker()
	results := make(chan workflow.Result, 1)

	go walker.ExecuteAsync(*playbook, results)

	select {
	case <-ctx.Done():
		return uuid.Nil, ctx.Err()
	case r := <-results:
		return r.RunId, nil
	}
}

// List returns all recorded run summaries.
func (s *Service) List(ctx context.Context) ([]runstate.RunEntry, error) {
	_ = ctx
	return s.reports.GetRuns()
}

// Report returns the detailed report for a single run.
func (s *Service) Report(ctx context.Context, runID uuid.UUID) (runstate.RunEntry, error) {
	_ = ctx
	return s.reports.GetRunReport(runID)
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

// DecodeVariables decodes a JSON run payload into variables.
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
