package decomposer_test

import (
	"errors"
	"fmt"
	"testing"

	"soarca/internal/decomposer"
	"soarca/models/cacao"
	"soarca/test/mocks/mock_executor"
	"soarca/test/mocks/mock_guid"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestExecutePlaybook(t *testing.T) {
	executer := new(mock_executor.Mock_Executor)
	uuid_mock := new(mock_guid.Mock_Guid)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variables{
		ObjectType: "string",
		Name:       "var1",
		Value:      "testing",
	}

	decomposer := decomposer.New(executer, uuid_mock)

	var step1 = cacao.Step{
		ObjectType:    "ssh",
		ID:            "action--test",
		Name:          "ssh-tests",
		StepVariables: map[string]cacao.Variables{expectedVariables.Name: expectedVariables},
		Commands:      []cacao.Command{expectedCommand},
		Cases:         map[string][]string{},
		OnCompletion:  "end--test",
		Agent:         "agent1",
		//OnCompletion:  "action--test2",
	}

	var end = cacao.Step{
		ObjectType: "end",
		ID:         "end--test",
		Name:       "end step",
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name: "user",
	}

	expectedTarget := cacao.Target{
		Name:               "sometarget",
		AuthInfoIdentifier: "id",
	}

	playbook := cacao.Playbook{
		ID:                            "test",
		Type:                          "test",
		Name:                          "ssh-test",
		WorkflowStart:                 "action--test",
		AuthenticationInfoDefinitions: map[string]cacao.AuthenticationInformation{"id": expectedAuth},
		AgentDefinitions:              map[string]cacao.Target{"agent1": expectedTarget},

		Workflow: map[string]cacao.Step{step1.ID: step1, end.ID: end},
	}

	var id, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	uuid_mock.On("New").Return(id)

	executer.On("Execute", id, expectedCommand,
		expectedAuth,
		expectedTarget,
		map[string]cacao.Variables{"var1": expectedVariables},
		"ssh").Return(id, map[string]cacao.Variables{}, nil)

	returnedId, err := decomposer.Execute(playbook)
	uuid_mock.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, nil)
	assert.Equal(t, returnedId.ExecutionId, id)
	executer.AssertExpectations(t)
}

func TestExecutePlaybookMultiStep(t *testing.T) {

	executer := new(mock_executor.Mock_Executor)
	uuid_mock := new(mock_guid.Mock_Guid)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedCommand2 := cacao.Command{
		Type:    "ssh2",
		Command: "ssh pwd",
	}

	expectedVariables := cacao.Variables{
		ObjectType: "string",
		Name:       "var1",
		Value:      "testing",
	}

	expectedVariables2 := cacao.Variables{
		ObjectType: "string",
		Name:       "var2",
		Value:      "testing2",
	}

	decomposer := decomposer.New(executer, uuid_mock)

	var step1 = cacao.Step{
		ObjectType:    "ssh",
		ID:            "action--test",
		Name:          "ssh-tests",
		StepVariables: map[string]cacao.Variables{expectedVariables.Name: expectedVariables},
		Commands:      []cacao.Command{expectedCommand},
		Cases:         map[string][]string{},
		OnCompletion:  "action--test2",
		Agent:         "agent1",
		//OnCompletion:  "action--test2",
	}

	var step2 = cacao.Step{
		ObjectType:    "ssh",
		ID:            "action--test2",
		Name:          "ssh-tests",
		StepVariables: map[string]cacao.Variables{expectedVariables2.Name: expectedVariables2},
		Commands:      []cacao.Command{expectedCommand2},
		Cases:         map[string][]string{},
		OnCompletion:  "end--test",
	}
	var step3 = cacao.Step{
		ObjectType:    "ssh",
		ID:            "action--test3",
		Name:          "ssh-tests",
		StepVariables: map[string]cacao.Variables{expectedVariables.Name: expectedVariables},
		Commands:      []cacao.Command{expectedCommand},
	}
	var end = cacao.Step{
		ObjectType: "end",
		ID:         "end--test",
		Name:       "end step",
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name: "user",
	}

	expectedTarget := cacao.Target{
		Name:               "sometarget",
		AuthInfoIdentifier: "id",
	}

	playbook := cacao.Playbook{
		ID:                            "test",
		Type:                          "test",
		Name:                          "ssh-test",
		WorkflowStart:                 "action--test",
		AuthenticationInfoDefinitions: map[string]cacao.AuthenticationInformation{"id": expectedAuth},
		AgentDefinitions:              map[string]cacao.Target{"agent1": expectedTarget},

		Workflow: map[string]cacao.Step{step1.ID: step1, step2.ID: step2, step3.ID: step3, end.ID: end},
	}

	var id, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	uuid_mock.On("New").Return(id)

	executer.On("Execute", id, expectedCommand,
		expectedAuth,
		expectedTarget,
		map[string]cacao.Variables{"var1": expectedVariables},
		"ssh").Return(id, map[string]cacao.Variables{}, nil)

	executer.On("Execute", id, expectedCommand2,
		cacao.AuthenticationInformation{},
		cacao.Target{},
		map[string]cacao.Variables{"var2": expectedVariables2},
		"ssh").Return(id, map[string]cacao.Variables{}, nil)

	returnedId, err := decomposer.Execute(playbook)
	uuid_mock.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, nil)
	assert.Equal(t, returnedId.ExecutionId, id)
	executer.AssertExpectations(t)

}

/*
Test with an Empty OnCompletion will result in not executing the step.
*/
func TestExecuteEmptyMultiStep(t *testing.T) {

	executer2 := new(mock_executor.Mock_Executor)
	uuid_mock2 := new(mock_guid.Mock_Guid)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variables{
		ObjectType: "string",
		Name:       "var1",
		Value:      "testing",
	}

	decomposer2 := decomposer.New(executer2, uuid_mock2)

	var step1 = cacao.Step{
		ObjectType:    "ssh",
		ID:            "action--test",
		Name:          "ssh-tests",
		StepVariables: map[string]cacao.Variables{expectedVariables.Name: expectedVariables},
		Commands:      []cacao.Command{expectedCommand},
		Cases:         map[string][]string{},
		OnCompletion:  "",
	}

	playbook := cacao.Playbook{
		ID:            "test",
		Type:          "test",
		Name:          "ssh-test",
		WorkflowStart: "action--test",

		Workflow: map[string]cacao.Step{step1.ID: step1},
	}

	var id, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	uuid_mock2.On("New").Return(id)

	returnedId, err := decomposer2.Execute(playbook)
	uuid_mock2.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, errors.New("empty on_completion_id"))
	assert.Equal(t, returnedId.ExecutionId, id)
	executer2.AssertExpectations(t)
}

/*
Test with an not occuring on completion id will result in not executing the step.
*/
func TestExecuteIllegalMultiStep(t *testing.T) {

	executer2 := new(mock_executor.Mock_Executor)
	uuid_mock2 := new(mock_guid.Mock_Guid)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variables{
		ObjectType: "string",
		Name:       "var1",
		Value:      "testing",
	}

	decomposer2 := decomposer.New(executer2, uuid_mock2)

	var step1 = cacao.Step{
		ObjectType:    "ssh",
		ID:            "action--test",
		Name:          "ssh-tests",
		StepVariables: map[string]cacao.Variables{expectedVariables.Name: expectedVariables},
		Commands:      []cacao.Command{expectedCommand},
		Cases:         map[string][]string{},
		OnCompletion:  "action-some-non-existing",
	}

	playbook := cacao.Playbook{
		ID:            "test",
		Type:          "test",
		Name:          "ssh-test",
		WorkflowStart: "action--test",

		Workflow: map[string]cacao.Step{step1.ID: step1},
	}

	var id, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	uuid_mock2.On("New").Return(id)

	returnedId, err := decomposer2.Execute(playbook)
	uuid_mock2.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, errors.New("on_completion_id key is not in workflows"))
	assert.Equal(t, returnedId.ExecutionId, id)
	executer2.AssertExpectations(t)
}
