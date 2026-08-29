package trigger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"soarca/internal/services"
	"soarca/internal/storage"
	"soarca/pkg/models/cacao"
)

// ValidationError marks trigger payload validation failures.
type ValidationError struct {
	Err error
}

func (e ValidationError) Error() string {
	return e.Err.Error()
}

func (e ValidationError) Unwrap() error {
	return e.Err
}

// Service implements the services.TriggerService interface.
type Service struct {
	executionRuntime services.ExecutionRuntime
	playbookStore    storage.PlaybookStore
}

// New creates a new trigger service.
func New(executionRuntime services.ExecutionRuntime, playbookStore storage.PlaybookStore) *Service {
	return &Service{
		executionRuntime: executionRuntime,
		playbookStore:    playbookStore,
	}
}

// ExecutePlaybook loads a stored playbook, applies trigger variables and starts execution.
func (s *Service) ExecutePlaybook(ctx context.Context, playbookID string, variables cacao.Variables) (executionID uuid.UUID, err error) {
	playbook, err := s.playbookStore.Get(ctx, playbookID)
	if err != nil {
		return uuid.Nil, err
	}
	if err := mergeVariablesInPlaybook(&playbook, variables); err != nil {
		return uuid.Nil, ValidationError{Err: err}
	}
	execID, err := s.executionRuntime.StartExecution(ctx, &playbook, cacao.Variables{})
	if err != nil {
		return uuid.Nil, err
	}
	return execID, nil
}

// ExecuteUploadedPlaybook executes a newly uploaded playbook after validation.
func (s *Service) ExecuteUploadedPlaybook(ctx context.Context, playbook *cacao.Playbook, variables cacao.Variables) (executionID uuid.UUID, err error) {
	if err := mergeVariablesInPlaybook(playbook, variables); err != nil {
		return uuid.Nil, ValidationError{Err: err}
	}
	execID, err := s.executionRuntime.StartExecution(ctx, playbook, cacao.Variables{})
	if err != nil {
		return uuid.Nil, err
	}
	return execID, nil
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
