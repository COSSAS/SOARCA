package cache_test

import (
	"errors"
	"soarca/internal/reporter/downstream_reporter/cache"
	"soarca/models/cacao"
	"soarca/models/report"
	"soarca/test/unittest/mocks/mock_utils/time"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestReportWorkflowStartFirst(t *testing.T) {

	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time)

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
	executionId0, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	expectedExecutionEntry := report.ExecutionEntry{
		ExecutionId: executionId0,
		PlaybookId:  "test",
		StepResults: map[string]report.StepResult{},
		Status:      report.Ongoing,
		Started:     timeNow,
		Ended:       time.Time{},
	}
	expectedExecutions := []string{"6ba7b810-9dad-11d1-80b4-00c04fd430c0"}

	_ = cacheReporter.ReportWorkflowStart(executionId0, playbook)
	exec, err := cacheReporter.GetExecutionReport(executionId0)
	assert.Equal(t, expectedExecutions, cacheReporter.GetExecutionsIDs())
	assert.Equal(t, expectedExecutionEntry.ExecutionId, exec.ExecutionId)
	assert.Equal(t, expectedExecutionEntry.PlaybookId, exec.PlaybookId)
	assert.Equal(t, expectedExecutionEntry.StepResults, exec.StepResults)
	assert.Equal(t, expectedExecutionEntry.Started, timeNow)
	assert.Equal(t, expectedExecutionEntry.Ended, time.Time{})
	assert.Equal(t, expectedExecutionEntry.Status, exec.Status)
	assert.Equal(t, err, nil)
	mock_time.AssertExpectations(t)
}

func TestReportWorkflowStartFifo(t *testing.T) {
	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time)

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
	executionId0, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	executionId1, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c1")
	executionId2, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c2")
	executionId3, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c3")
	executionId4, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c4")
	executionId5, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c5")
	executionId6, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c6")
	executionId7, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c7")
	executionId8, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	executionId9, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c9")
	executionId10, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430ca")

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	expectedExecutionsFull := []string{
		"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c1",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c2",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c3",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c4",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c5",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c6",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c7",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c9",
	}
	expectedExecutionsFifo := []string{
		"6ba7b810-9dad-11d1-80b4-00c04fd430c1",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c2",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c3",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c4",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c5",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c6",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c7",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c9",
		"6ba7b810-9dad-11d1-80b4-00c04fd430ca",
	}

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId1, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId2, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId3, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId4, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId5, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId6, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId7, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId8, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId9, playbook)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, expectedExecutionsFull, cacheReporter.GetExecutionsIDs())

	err = cacheReporter.ReportWorkflowStart(executionId10, playbook)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, expectedExecutionsFifo, cacheReporter.GetExecutionsIDs())
	mock_time.AssertExpectations(t)
}

func TestReportWorkflowEnd(t *testing.T) {

	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time)

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
	executionId0, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	expectedExecutions := []string{"6ba7b810-9dad-11d1-80b4-00c04fd430c0"}

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowEnd(executionId0, playbook, nil)
	if err != nil {
		t.Fail()
	}

	expectedExecutionEntry := report.ExecutionEntry{
		ExecutionId: executionId0,
		PlaybookId:  "test",
		Started:     timeNow,
		Ended:       timeNow,
		StepResults: map[string]report.StepResult{},
		Status:      report.SuccessfullyExecuted,
	}

	exec, err := cacheReporter.GetExecutionReport(executionId0)
	assert.Equal(t, expectedExecutions, cacheReporter.GetExecutionsIDs())
	assert.Equal(t, expectedExecutionEntry.ExecutionId, exec.ExecutionId)
	assert.Equal(t, expectedExecutionEntry.PlaybookId, exec.PlaybookId)
	assert.Equal(t, expectedExecutionEntry.StepResults, exec.StepResults)
	assert.Equal(t, expectedExecutionEntry.Status, exec.Status)
	assert.Equal(t, exec.Ended, expectedExecutionEntry.Ended)
	assert.Equal(t, err, nil)
	mock_time.AssertExpectations(t)
}

func TestReportStepStartAndEnd(t *testing.T) {
	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time)

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
	executionId0, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepStart(executionId0, step1, cacao.NewVariables(expectedVariables))
	if err != nil {
		t.Fail()
	}

	expectedStepStatus := report.StepResult{
		ExecutionId: executionId0,
		StepId:      step1.ID,
		Started:     timeNow,
		Ended:       time.Time{},
		Variables:   cacao.NewVariables(expectedVariables),
		Status:      report.Ongoing,
		Error:       nil,
	}

	exec, err := cacheReporter.GetExecutionReport(executionId0)
	stepStatus := exec.StepResults[step1.ID]
	assert.Equal(t, stepStatus.ExecutionId, expectedStepStatus.ExecutionId)
	assert.Equal(t, stepStatus.StepId, expectedStepStatus.StepId)
	assert.Equal(t, stepStatus.Started, expectedStepStatus.Started)
	assert.Equal(t, stepStatus.Ended, expectedStepStatus.Ended)
	assert.Equal(t, stepStatus.Variables, expectedStepStatus.Variables)
	assert.Equal(t, stepStatus.Status, expectedStepStatus.Status)
	assert.Equal(t, stepStatus.Error, expectedStepStatus.Error)
	assert.Equal(t, err, nil)

	err = cacheReporter.ReportStepEnd(executionId0, step1, cacao.NewVariables(expectedVariables), nil)
	if err != nil {
		t.Fail()
	}

	expectedStepResult := report.StepResult{
		ExecutionId: executionId0,
		StepId:      step1.ID,
		Started:     timeNow,
		Ended:       timeNow,
		Variables:   cacao.NewVariables(expectedVariables),
		Status:      report.SuccessfullyExecuted,
		Error:       nil,
	}

	exec, err = cacheReporter.GetExecutionReport(executionId0)
	stepResult := exec.StepResults[step1.ID]
	assert.Equal(t, stepResult.ExecutionId, expectedStepResult.ExecutionId)
	assert.Equal(t, stepResult.StepId, expectedStepResult.StepId)
	assert.Equal(t, stepResult.Started, expectedStepResult.Started)
	assert.Equal(t, stepResult.Ended, expectedStepResult.Ended)
	assert.Equal(t, stepResult.Variables, expectedStepResult.Variables)
	assert.Equal(t, stepResult.Status, expectedStepResult.Status)
	assert.Equal(t, stepResult.Error, expectedStepResult.Error)
	assert.Equal(t, err, nil)
	mock_time.AssertExpectations(t)
}

func TestInvalidStepReportAfterExecutionEnd(t *testing.T) {
	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time)

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
	executionId0, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepStart(executionId0, step1, cacao.NewVariables(expectedVariables))
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowEnd(executionId0, playbook, nil)
	if err != nil {
		t.Fail()
	}

	err = cacheReporter.ReportStepEnd(executionId0, step1, cacao.NewVariables(expectedVariables), nil)
	if err == nil {
		t.Fail()
	}

	expectedErr := errors.New("trying to report on the execution of a step for an already reported completed or failed execution")
	assert.Equal(t, err, expectedErr)
	mock_time.AssertExpectations(t)
}

func TestInvalidStepReportAfterStepEnd(t *testing.T) {
	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time)

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
	executionId0, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepStart(executionId0, step1, cacao.NewVariables(expectedVariables))
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepEnd(executionId0, step1, cacao.NewVariables(expectedVariables), nil)
	if err != nil {
		t.Fail()
	}

	err = cacheReporter.ReportStepEnd(executionId0, step1, cacao.NewVariables(expectedVariables), nil)
	if err == nil {
		t.Fail()
	}

	expectedErr := errors.New("trying to report on the execution of a step that was already reported completed or failed")
	assert.Equal(t, err, expectedErr)
	mock_time.AssertExpectations(t)
}
