package decomposer_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"soarca/internal/decomposer"
	decomposer_model "soarca/internal/decomposer"
	"soarca/internal/executors/action"
	"soarca/models/cacao"
	"soarca/models/execution"
	"soarca/test/unittest/mocks/mock_executor"
	mock_condition_executor "soarca/test/unittest/mocks/mock_executor/condition"
	mock_playbook_action_executor "soarca/test/unittest/mocks/mock_executor/playbook_action"
	"soarca/test/unittest/mocks/mock_guid"
	"soarca/test/unittest/mocks/mock_reporter"
	mock_time "soarca/test/unittest/mocks/mock_utils/time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestExecutePlaybook(t *testing.T) {
	mock_action_executor := new(mock_executor.Mock_Action_Executor)
	mock_playbook_action_executor := new(mock_playbook_action_executor.Mock_PlaybookActionExecutor)
	mock_condition_executor := new(mock_condition_executor.Mock_Condition)
	uuid_mock := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

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
		mock_condition_executor,
		uuid_mock,
		mock_reporter,
		mock_time)

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
		Delay:         10,
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

	mock_reporter.On("ReportWorkflowStart", executionId, playbook).Return()
	mock_time.On("Sleep", time.Second*10).Return()
	mock_reporter.On("ReportWorkflowEnd", executionId, playbook, nil).Return()
	mock_action_executor.On("Execute", metaStep1, playbookStepMetadata).Return(cacao.NewVariables(cacao.Variable{Name: "return", Value: "value"}), nil)

	var nilch chan decomposer_model.ExecutionDetails
	details, err := decomposer.Execute(playbook, nilch)
	uuid_mock.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, nil)
	assert.Equal(t, details.ExecutionId, executionId)
	mock_action_executor.AssertExpectations(t)
	mock_reporter.AssertExpectations(t)
	mock_time.AssertExpectations(t)
	value, found := details.Variables.Find("return")
	assert.Equal(t, found, true)
	assert.Equal(t, value.Value, "value")
}

func TestExecutePlaybookMultiStep(t *testing.T) {
	mock_action_executor := new(mock_executor.Mock_Action_Executor)
	mock_playbook_action_executor := new(mock_playbook_action_executor.Mock_PlaybookActionExecutor)
	mock_condition_executor := new(mock_condition_executor.Mock_Condition)
	uuid_mock := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

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
		mock_condition_executor,
		uuid_mock,
		mock_reporter,
		mock_time)

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

	mock_reporter.On("ReportWorkflowStart", executionId, playbook).Return()
	mock_time.On("Sleep", time.Second*0).Return()
	mock_reporter.On("ReportWorkflowEnd", executionId, playbook, nil).Return()
	mock_action_executor.On("Execute", metaStep1, playbookStepMetadata1).Return(cacao.NewVariables(firstResult), nil)

	playbookStepMetadata2 := action.PlaybookStepMetadata{
		Step:      step2,
		Targets:   playbook.TargetDefinitions,
		Auth:      playbook.AuthenticationInfoDefinitions,
		Agent:     expectedAgent,
		Variables: cacao.NewVariables(expectedVariables2, firstResult),
	}

	mock_action_executor.On("Execute", metaStep2, playbookStepMetadata2).Return(cacao.NewVariables(cacao.Variable{Name: "result", Value: "updated"}), nil)

	var nilch chan decomposer_model.ExecutionDetails
	details, err := decomposer.Execute(playbook, nilch)
	uuid_mock.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, nil)
	assert.Equal(t, details.ExecutionId, executionId)
	mock_action_executor.AssertExpectations(t)
	mock_reporter.AssertExpectations(t)

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
	mock_condition_executor := new(mock_condition_executor.Mock_Condition)
	uuid_mock2 := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

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
		mock_condition_executor,
		uuid_mock2,
		mock_reporter,
		mock_time)

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

	mock_reporter.On("ReportWorkflowStart", id, playbook).Return()
	mock_time.On("Sleep", time.Second*0).Return()
	mock_reporter.On("ReportWorkflowEnd", id, playbook, errors.New("empty success step")).Return()

	var nilch chan decomposer_model.ExecutionDetails
	returnedId, err := decomposer2.Execute(playbook, nilch)
	uuid_mock2.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, errors.New("empty success step"))
	assert.Equal(t, returnedId.ExecutionId, id)
	mock_action_executor2.AssertExpectations(t)
	mock_reporter.AssertExpectations(t)
}

/*
Test with an not occuring on completion id will result in not executing the step.
*/
func TestExecuteIllegalMultiStep(t *testing.T) {
	mock_action_executor2 := new(mock_executor.Mock_Action_Executor)
	mock_playbook_action_executor2 := new(mock_playbook_action_executor.Mock_PlaybookActionExecutor)
	mock_condition_executor := new(mock_condition_executor.Mock_Condition)
	uuid_mock2 := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

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
		mock_condition_executor,
		uuid_mock2,
		mock_reporter,
		mock_time)

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
	mock_reporter.On("ReportWorkflowStart", id, playbook).Return()
	mock_time.On("Sleep", time.Second*0).Return()
	mock_reporter.On("ReportWorkflowEnd", id, playbook, errors.New("empty success step")).Return()

	var nilch chan decomposer_model.ExecutionDetails
	returnedId, err := decomposer2.Execute(playbook, nilch)
	uuid_mock2.AssertExpectations(t)
	mock_reporter.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, errors.New("empty success step"))
	assert.Equal(t, returnedId.ExecutionId, id)
	mock_action_executor2.AssertExpectations(t)
}

func TestExecutePlaybookAction(t *testing.T) {
	mock_action_executor := new(mock_executor.Mock_Action_Executor)
	mock_playbook_action_executor := new(mock_playbook_action_executor.Mock_PlaybookActionExecutor)
	mock_condition_executor := new(mock_condition_executor.Mock_Condition)
	uuid_mock := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)
	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	decomposer := decomposer.New(mock_action_executor,
		mock_playbook_action_executor,
		mock_condition_executor,
		uuid_mock,
		mock_reporter,
		mock_time)

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
	mock_reporter.On("ReportWorkflowStart", executionId, playbook).Return()
	mock_time.On("Sleep", time.Second*0).Return()
	mock_reporter.On("ReportWorkflowEnd", executionId, playbook, nil).Return()

	mock_playbook_action_executor.On("Execute",
		metaStep1,
		step1,
		cacao.NewVariables(expectedVariables)).Return(cacao.NewVariables(cacao.Variable{Name: "return", Value: "value"}), nil)

	var nilch chan decomposer_model.ExecutionDetails
	details, err := decomposer.Execute(playbook, nilch)
	uuid_mock.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, nil)
	assert.Equal(t, details.ExecutionId, executionId)
	mock_reporter.AssertExpectations(t)
	mock_action_executor.AssertExpectations(t)
	value, found := details.Variables.Find("return")
	assert.Equal(t, found, true)
	assert.Equal(t, value.Value, "value")
}

func TestExecuteIfCondition(t *testing.T) {

	mock_action_executor := new(mock_executor.Mock_Action_Executor)
	mock_playbook_action_executor := new(mock_playbook_action_executor.Mock_PlaybookActionExecutor)
	mock_condition_executor := new(mock_condition_executor.Mock_Condition)
	uuid_mock := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)
	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "__var1__",
		Value: "testing",
	}

	// returned from step
	expectedVariables2 := cacao.Variable{
		Type:  "string",
		Name:  "__var2__",
		Value: "testing2",
	}

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedTarget := cacao.AgentTarget{
		Name:               "sometarget",
		AuthInfoIdentifier: "id",
	}

	expectedAgent := cacao.AgentTarget{
		Type: "soarca",
		Name: "soarca-ssh",
	}

	decomposer := decomposer.New(mock_action_executor,
		mock_playbook_action_executor,
		mock_condition_executor,
		uuid_mock,
		mock_reporter,
		mock_time)

	end := cacao.Step{
		Type: cacao.StepTypeEnd,
		ID:   "end--test",
		Name: "end step",
	}

	endTrue := cacao.Step{
		Type: cacao.StepTypeEnd,
		ID:   "end--true",
		Name: "end branch true step",
	}

	stepTrue := cacao.Step{
		Type:          cacao.StepTypeAction,
		ID:            "action--step-true",
		Name:          "ssh-tests",
		Commands:      []cacao.Command{expectedCommand},
		Targets:       []string{expectedTarget.ID},
		StepVariables: cacao.NewVariables(expectedVariables),
		OnCompletion:  endTrue.ID,
	}

	endFalse := cacao.Step{
		Type: cacao.StepTypeEnd,
		ID:   "end--false",
		Name: "end branch false step",
	}

	stepFalse := cacao.Step{
		Type:          cacao.StepTypeAction,
		ID:            "action--step-false",
		Name:          "ssh-tests",
		Commands:      []cacao.Command{expectedCommand},
		Targets:       []string{expectedTarget.ID},
		StepVariables: cacao.NewVariables(expectedVariables),
		OnCompletion:  endFalse.ID,
	}

	stepCompletion := cacao.Step{
		Type:          cacao.StepTypeAction,
		ID:            "action--step-completion",
		Name:          "ssh-tests",
		Commands:      []cacao.Command{expectedCommand},
		Targets:       []string{expectedTarget.ID},
		StepVariables: cacao.NewVariables(expectedVariables),
		OnCompletion:  end.ID,
	}

	stepIf := cacao.Step{
		Type:          cacao.StepTypeIfCondition,
		ID:            "if-condition--test",
		Name:          "if condition",
		StepVariables: cacao.NewVariables(expectedVariables),
		Condition:     "__var1__:value = testing",
		OnTrue:        stepTrue.ID,
		OnFalse:       stepFalse.ID,
		OnCompletion:  stepCompletion.ID,
	}

	start := cacao.Step{
		Type:         cacao.StepTypeStart,
		ID:           "start--test",
		Name:         "start step",
		OnCompletion: stepIf.ID,
	}

	playbook := cacao.Playbook{
		ID:            "test",
		Type:          "test",
		Name:          "playbook-test",
		WorkflowStart: start.ID,
		Workflow: map[string]cacao.Step{start.ID: start,
			stepIf.ID:         stepIf,
			stepTrue.ID:       stepTrue,
			stepFalse.ID:      stepFalse,
			stepCompletion.ID: stepCompletion,
			end.ID:            end,
			endTrue.ID:        endTrue,
			endFalse.ID:       endFalse},
		AgentDefinitions:  map[string]cacao.AgentTarget{expectedAgent.ID: expectedAgent},
		TargetDefinitions: map[string]cacao.AgentTarget{expectedTarget.ID: expectedTarget},
	}

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	metaStepIf := execution.Metadata{ExecutionId: executionId, PlaybookId: "test", StepId: stepIf.ID}

	uuid_mock.On("New").Return(executionId)
	mock_reporter.On("ReportWorkflowStart", executionId, playbook).Return()
	mock_time.On("Sleep", time.Second*0).Return()

	mock_condition_executor.On("Execute",
		metaStepIf,
		stepIf,
		cacao.NewVariables(expectedVariables)).Return(stepTrue.ID, true, nil)

	stepTrueDetails := action.PlaybookStepMetadata{
		Step:      stepTrue,
		Targets:   playbook.TargetDefinitions,
		Auth:      playbook.AuthenticationInfoDefinitions,
		Agent:     expectedAgent,
		Variables: cacao.NewVariables(expectedVariables),
	}

	metaStepTrue := execution.Metadata{ExecutionId: executionId, PlaybookId: "test", StepId: stepTrue.ID}
	mock_time.On("Sleep", time.Second*0).Return()

	mock_action_executor.On("Execute",
		metaStepTrue,
		stepTrueDetails).Return(cacao.NewVariables(expectedVariables2), nil)

	stepCompletionDetails := action.PlaybookStepMetadata{
		Step:      stepCompletion,
		Targets:   playbook.TargetDefinitions,
		Auth:      playbook.AuthenticationInfoDefinitions,
		Agent:     expectedAgent,
		Variables: cacao.NewVariables(expectedVariables, expectedVariables2),
	}

	metaStepCompletion := execution.Metadata{ExecutionId: executionId, PlaybookId: "test", StepId: stepCompletion.ID}
	mock_time.On("Sleep", time.Second*0).Return()

	mock_action_executor.On("Execute",
		metaStepCompletion,
		stepCompletionDetails).Return(cacao.NewVariables(), nil)
	mock_reporter.On("ReportWorkflowEnd", executionId, playbook, nil).Return()
	var nilch chan decomposer_model.ExecutionDetails
	details, err := decomposer.Execute(playbook, nilch)
	uuid_mock.AssertExpectations(t)
	fmt.Println(err)
	assert.Equal(t, err, nil)
	assert.Equal(t, details.ExecutionId, executionId)
	mock_reporter.AssertExpectations(t)
	mock_condition_executor.AssertExpectations(t)
	mock_action_executor.AssertExpectations(t)

}

func TestDelayStepExecution(t *testing.T) {
	mock_action_executor := new(mock_executor.Mock_Action_Executor)
	mock_playbook_action_executor := new(mock_playbook_action_executor.Mock_PlaybookActionExecutor)
	mock_condition_executor := new(mock_condition_executor.Mock_Condition)
	uuid_mock := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

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
		Delay:         10,
	}

	decomposer := decomposer.New(mock_action_executor,
		mock_playbook_action_executor,
		mock_condition_executor,
		uuid_mock,
		mock_reporter,
		mock_time)

	executionId, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")
	metaStep1 := execution.Metadata{ExecutionId: executionId, PlaybookId: "", StepId: step1.ID}
	playbookStepMetadata := action.PlaybookStepMetadata{
		Step:      step1,
		Variables: cacao.NewVariables(expectedVariables),
	}

	mock_time.On("Sleep", time.Second*10).Return()
	mock_action_executor.On("Execute", metaStep1, playbookStepMetadata).Return(cacao.NewVariables(cacao.Variable{Name: "return", Value: "value"}), nil)

	_, err := decomposer.ExecuteStep(step1, cacao.NewVariables(expectedVariables))
	assert.Equal(t, err, nil)

}

func TestDelayStepNegativeTimeExecution(t *testing.T) {
	mock_action_executor := new(mock_executor.Mock_Action_Executor)
	mock_playbook_action_executor := new(mock_playbook_action_executor.Mock_PlaybookActionExecutor)
	mock_condition_executor := new(mock_condition_executor.Mock_Condition)
	uuid_mock := new(mock_guid.Mock_Guid)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

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
		Delay:         -10,
	}

	decomposer := decomposer.New(mock_action_executor,
		mock_playbook_action_executor,
		mock_condition_executor,
		uuid_mock,
		mock_reporter,
		mock_time)

	executionId, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")
	metaStep1 := execution.Metadata{ExecutionId: executionId, PlaybookId: "", StepId: step1.ID}
	playbookStepMetadata := action.PlaybookStepMetadata{
		Step:      step1,
		Variables: cacao.NewVariables(expectedVariables),
	}

	mock_time.On("Sleep", time.Second*-10).Return()
	mock_action_executor.On("Execute", metaStep1, playbookStepMetadata).Return(cacao.NewVariables(cacao.Variable{Name: "return", Value: "value"}), nil)

	_, err := decomposer.ExecuteStep(step1, cacao.NewVariables(expectedVariables))
	assert.Equal(t, err, nil)

}
