package decomposer_test

import (
	"fmt"
	"soarca/internal/decomposer"
	"soarca/models/cacao"
	"soarca/models/execution"
	"soarca/test/unittest/mocks/mock_executor"
	"soarca/test/unittest/mocks/mock_guid"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestParallelStep(t *testing.T) {
	mock_executer := new(mock_executor.Mock_Executor)
	uuid_mock := new(mock_guid.Mock_Guid)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	decomposer := decomposer.New(mock_executer, uuid_mock)

	end := cacao.Step{
		Type: "end",
		ID:   "end--test",
		Name: "end step",
	}

	step1 := cacao.Step{
		Type:         "action",
		ID:           "action--test",
		Name:         "http-tests",
		Commands:     []cacao.Command{expectedCommand},
		Cases:        map[string]string{},
		OnCompletion: "end--test",
		Agent:        "agent1",
		Targets:      []string{"target1"},
	}

	step2 := cacao.Step{
		Type:         "action",
		ID:           "action2--test",
		Name:         "http-tests",
		Commands:     []cacao.Command{expectedCommand},
		Cases:        map[string]string{},
		OnCompletion: "end--test",
		Agent:        "agent1",
		Targets:      []string{"target1"},
	}

	parallelStep := cacao.Step{
		Type:         "parallel",
		ID:           "parallel--test",
		OnCompletion: end.ID,
		NextSteps:    []string{step1.ID, step2.ID},
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name: "user",
	}

	expectedTarget := cacao.AgentTarget{
		Name:               "sometarget",
		AuthInfoIdentifier: "id",
	}

	expectedAgent := cacao.AgentTarget{
		Type: "orches",
		Name: "soarca-ssh",
	}

	playbook := cacao.Playbook{
		ID:                            "test",
		Type:                          "test",
		Name:                          "ssh-test",
		WorkflowStart:                 parallelStep.ID,
		AuthenticationInfoDefinitions: map[string]cacao.AuthenticationInformation{"id": expectedAuth},
		AgentDefinitions:              map[string]cacao.AgentTarget{"agent1": expectedAgent},
		TargetDefinitions:             map[string]cacao.AgentTarget{"target1": expectedTarget},

		Workflow: map[string]cacao.Step{
			parallelStep.ID: parallelStep,
			step1.ID:        step1,
			step2.ID:        step2,
			end.ID:          end,
		},
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	metaStep1 := execution.Metadata{ExecutionId: executionId, PlaybookId: "test", StepId: step1.ID}
	metaStep2 := execution.Metadata{ExecutionId: executionId, PlaybookId: "test", StepId: step2.ID}

	uuid_mock.On("New").Return(executionId)

	mock_executer.On("Execute",
		metaStep1,
		expectedCommand,
		expectedAuth,
		expectedTarget,
		cacao.Variables{},
		expectedAgent).Return(executionId, cacao.Variables{}, nil)

	mock_executer.On("Execute",
		metaStep2,
		expectedCommand,
		expectedAuth,
		expectedTarget,
		cacao.Variables{},
		expectedAgent).Return(executionId, cacao.Variables{}, nil)

	returnedId, err := decomposer.Execute(playbook)
	uuid_mock.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, nil)
	assert.Equal(t, returnedId.ExecutionId, executionId)
	mock_executer.AssertExpectations(t)
}
