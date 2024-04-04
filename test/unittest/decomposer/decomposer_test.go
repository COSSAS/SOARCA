package decomposer_test

import (
	"errors"
	"fmt"
	"testing"

	"soarca/internal/decomposer"
	"soarca/internal/executors/action"
	"soarca/models/cacao"
	"soarca/models/execution"
	"soarca/test/unittest/mocks/mock_executor"
	mock_playbook_action_executor "soarca/test/unittest/mocks/mock_executor/playbook_action"
	"soarca/test/unittest/mocks/mock_guid"
	"soarca/test/unittest/mocks/mock_reporter"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestExecutePlaybook(t *testing.T) {
	mock_action_executor := new(mock_executor.Mock_Action_Executor)
	mock_playbook_action_executor := new(mock_playbook_action_executor.Mock_PlaybookActionExecutor)
	uuid_mock := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	decomposer := decomposer.New(mock_action_executor,
		mock_playbook_action_executor,
		uuid_mock, mock_reporter)

	step1 := cacao.Step{
		Type:          "action",
		ID:            "action--test",
		Name:          "ssh-tests",
		StepVariables: cacao.NewVariables(expectedVariables),
		Commands:      []cacao.Command{expectedCommand},
		Cases:         map[string]string{},
		OnCompletion:  "end--test",
		Agent:         "agent1",
		Targets:       []string{"target1"},
	}

	end := cacao.Step{
		Type: "end",
		ID:   "end--test",
		Name: "end step",
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name: "user",
		ID:   "auth1",
	}

	expectedTarget := cacao.AgentTarget{
		Name:               "sometarget",
		AuthInfoIdentifier: "auth1",
		ID:                 "target1",
	}

	expectedAgent := cacao.AgentTarget{
		Type: "soarca",
		Name: "soarca-ssh",
	}

	playbook := cacao.Playbook{
		ID:                            "test",
		Type:                          "test",
		Name:                          "ssh-test",
		WorkflowStart:                 step1.ID,
		AuthenticationInfoDefinitions: map[string]cacao.AuthenticationInformation{"id": expectedAuth},
		AgentDefinitions:              map[string]cacao.AgentTarget{"agent1": expectedAgent},
		TargetDefinitions:             map[string]cacao.AgentTarget{"target1": expectedTarget},

		Workflow: map[string]cacao.Step{step1.ID: step1, end.ID: end},
	}

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	metaStep1 := execution.Metadata{ExecutionId: executionId, PlaybookId: "test", StepId: step1.ID}

	uuid_mock.On("New").Return(executionId)

	playbookStepMetadata := action.PlaybookStepMetadata{
		Step:      step1,
		Targets:   playbook.TargetDefinitions,
		Auth:      playbook.AuthenticationInfoDefinitions,
		Agent:     expectedAgent,
		Variables: cacao.NewVariables(expectedVariables),
	}

	mock_action_executor.On("Execute", metaStep1, playbookStepMetadata).Return(cacao.NewVariables(cacao.Variable{Name: "return", Value: "value"}), nil)

	details, err := decomposer.Execute(playbook)
	uuid_mock.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, nil)
	assert.Equal(t, details.ExecutionId, executionId)
	mock_action_executor.AssertExpectations(t)
	value, found := details.Variables.Find("return")
	assert.Equal(t, found, true)
	assert.Equal(t, value.Value, "value")
}

func TestExecutePlaybookMultiStep(t *testing.T) {
	mock_action_executor := new(mock_executor.Mock_Action_Executor)
	mock_playbook_action_executor := new(mock_playbook_action_executor.Mock_PlaybookActionExecutor)
	uuid_mock := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedCommand2 := cacao.Command{
		Type:    "ssh2",
		Command: "ssh pwd",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	expectedVariables2 := cacao.Variable{
		Type:  "string",
		Name:  "var2",
		Value: "testing2",
	}

	decomposer := decomposer.New(mock_action_executor,
		mock_playbook_action_executor,
		uuid_mock, mock_reporter)

	step1 := cacao.Step{
		Type:          "action",
		ID:            "action--test",
		Name:          "ssh-tests",
		StepVariables: cacao.NewVariables(expectedVariables),
		Commands:      []cacao.Command{expectedCommand},
		Cases:         map[string]string{},
		OnCompletion:  "action--test2",
		Agent:         "agent1",
		Targets:       []string{"target1"},
		// OnCompletion:  "action--test2",
	}

	step2 := cacao.Step{
		Type:          "action",
		ID:            "action--test2",
		Name:          "ssh-tests",
		StepVariables: cacao.NewVariables(expectedVariables2),
		Commands:      []cacao.Command{expectedCommand2},
		Cases:         map[string]string{},
		Agent:         "agent1",
		Targets:       []string{"target1"},
		OnCompletion:  "end--test",
	}
	step3 := cacao.Step{
		Type:          "action",
		ID:            "action--test3",
		Name:          "ssh-tests",
		StepVariables: cacao.NewVariables(expectedVariables2),
		Commands:      []cacao.Command{expectedCommand},
		Agent:         "agent1",
		// Targets:       []string{"target1"},
	}
	end := cacao.Step{
		Type: "end",
		ID:   "end--test",
		Name: "end step",
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name: "user",
		ID:   "auth1",
	}

	expectedTarget := cacao.AgentTarget{
		Name:               "sometarget",
		AuthInfoIdentifier: "auth1",
		ID:                 "target1",
	}

	expectedAgent := cacao.AgentTarget{
		Type: "soarca",
		Name: "soarca-ssh",
	}

	playbook := cacao.Playbook{
		ID:                            "test",
		Type:                          "test",
		Name:                          "ssh-test",
		WorkflowStart:                 "action--test",
		AuthenticationInfoDefinitions: map[string]cacao.AuthenticationInformation{"id": expectedAuth},
		AgentDefinitions:              map[string]cacao.AgentTarget{"agent1": expectedAgent},
		TargetDefinitions:             map[string]cacao.AgentTarget{"target1": expectedTarget},

		Workflow: map[string]cacao.Step{step1.ID: step1, step2.ID: step2, step3.ID: step3, end.ID: end},
	}

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	metaStep1 := execution.Metadata{ExecutionId: executionId, PlaybookId: "test", StepId: step1.ID}
	metaStep2 := execution.Metadata{ExecutionId: executionId, PlaybookId: "test", StepId: step2.ID}

	uuid_mock.On("New").Return(executionId)

	firstResult := cacao.Variable{Name: "result", Value: "value"}

	playbookStepMetadata1 := action.PlaybookStepMetadata{
		Step:      step1,
		Targets:   playbook.TargetDefinitions,
		Auth:      playbook.AuthenticationInfoDefinitions,
		Agent:     expectedAgent,
		Variables: cacao.NewVariables(expectedVariables),
	}

	mock_action_executor.On("Execute", metaStep1, playbookStepMetadata1).Return(cacao.NewVariables(firstResult), nil)

	playbookStepMetadata2 := action.PlaybookStepMetadata{
		Step:      step2,
		Targets:   playbook.TargetDefinitions,
		Auth:      playbook.AuthenticationInfoDefinitions,
		Agent:     expectedAgent,
		Variables: cacao.NewVariables(expectedVariables2, firstResult),
	}

	mock_action_executor.On("Execute", metaStep2, playbookStepMetadata2).Return(cacao.NewVariables(cacao.Variable{Name: "result", Value: "updated"}), nil)

	details, err := decomposer.Execute(playbook)
	uuid_mock.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, nil)
	assert.Equal(t, details.ExecutionId, executionId)
	mock_action_executor.AssertExpectations(t)

	value, found := details.Variables.Find("result")
	assert.Equal(t, found, true)
	assert.Equal(t, value.Value, "updated") // Value overwritten
}

/*
Test with an Empty OnCompletion will result in not executing the step.
*/
func TestExecuteEmptyMultiStep(t *testing.T) {
	mock_action_executor2 := new(mock_executor.Mock_Action_Executor)
	mock_playbook_action_executor2 := new(mock_playbook_action_executor.Mock_PlaybookActionExecutor)
	uuid_mock2 := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	expectedTarget := cacao.AgentTarget{
		Name:               "sometarget",
		AuthInfoIdentifier: "id",
	}

	expectedAgent := cacao.AgentTarget{
		Type: "soarca",
		Name: "soarca-ssh",
	}

	decomposer2 := decomposer.New(mock_action_executor2,
		mock_playbook_action_executor2,
		uuid_mock2, mock_reporter)

	step1 := cacao.Step{
		Type:          "ssh",
		ID:            "action--test",
		Name:          "ssh-tests",
		Agent:         "agent1",
		StepVariables: cacao.NewVariables(expectedVariables),
		Commands:      []cacao.Command{expectedCommand},
		Cases:         map[string]string{},
		OnCompletion:  "", // Empty
	}

	playbook := cacao.Playbook{
		ID:                "test",
		Type:              "test",
		Name:              "ssh-test",
		WorkflowStart:     "action--test",
		AgentDefinitions:  map[string]cacao.AgentTarget{"agent1": expectedAgent},
		TargetDefinitions: map[string]cacao.AgentTarget{"target1": expectedTarget},
		Workflow:          map[string]cacao.Step{step1.ID: step1},
	}

	id, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	uuid_mock2.On("New").Return(id)

	returnedId, err := decomposer2.Execute(playbook)
	uuid_mock2.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, errors.New("empty success step"))
	assert.Equal(t, returnedId.ExecutionId, id)
	mock_action_executor2.AssertExpectations(t)
}

/*
Test with an not occuring on completion id will result in not executing the step.
*/
func TestExecuteIllegalMultiStep(t *testing.T) {
	mock_action_executor2 := new(mock_executor.Mock_Action_Executor)
	mock_playbook_action_executor2 := new(mock_playbook_action_executor.Mock_PlaybookActionExecutor)
	uuid_mock2 := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	decomposer2 := decomposer.New(mock_action_executor2,
		mock_playbook_action_executor2,
		uuid_mock2, mock_reporter)

	step1 := cacao.Step{
		Type:          "action",
		ID:            "action--test",
		Name:          "ssh-tests",
		StepVariables: cacao.NewVariables(expectedVariables),
		Commands:      []cacao.Command{expectedCommand},
		Cases:         map[string]string{},
		OnCompletion:  "action-some-non-existing",
	}

	playbook := cacao.Playbook{
		ID:            "test",
		Type:          "test",
		Name:          "ssh-test",
		WorkflowStart: "action--test",

		Workflow: map[string]cacao.Step{step1.ID: step1},
	}

	id, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	uuid_mock2.On("New").Return(id)

	returnedId, err := decomposer2.Execute(playbook)
	uuid_mock2.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, errors.New("empty success step"))
	assert.Equal(t, returnedId.ExecutionId, id)
	mock_action_executor2.AssertExpectations(t)
}

func TestExecutePlaybookAction(t *testing.T) {
	mock_action_executor := new(mock_executor.Mock_Action_Executor)
	mock_playbook_action_executor := new(mock_playbook_action_executor.Mock_PlaybookActionExecutor)
	uuid_mock := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	decomposer := decomposer.New(mock_action_executor,
		mock_playbook_action_executor,
		uuid_mock, mock_reporter)

	step1 := cacao.Step{
		Type:          "playbook-action",
		ID:            "playbook-action--test",
		Name:          "ssh-tests",
		StepVariables: cacao.NewVariables(expectedVariables),
		PlaybookID:    "playbook--1",
		OnCompletion:  "end--test",
	}

	end := cacao.Step{
		Type: "end",
		ID:   "end--test",
		Name: "end step",
	}

	playbook := cacao.Playbook{
		ID:            "test",
		Type:          "test",
		Name:          "playbook-test",
		WorkflowStart: step1.ID,
		Workflow:      map[string]cacao.Step{step1.ID: step1, end.ID: end},
	}

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	metaStep1 := execution.Metadata{ExecutionId: executionId, PlaybookId: "test", StepId: step1.ID}

	uuid_mock.On("New").Return(executionId)

	mock_playbook_action_executor.On("Execute",
		metaStep1,
		step1,
		cacao.NewVariables(expectedVariables)).Return(cacao.NewVariables(cacao.Variable{Name: "return", Value: "value"}), nil)

	details, err := decomposer.Execute(playbook)
	uuid_mock.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, nil)
	assert.Equal(t, details.ExecutionId, executionId)
	mock_action_executor.AssertExpectations(t)
	value, found := details.Variables.Find("return")
	assert.Equal(t, found, true)
	assert.Equal(t, value.Value, "value")
}
