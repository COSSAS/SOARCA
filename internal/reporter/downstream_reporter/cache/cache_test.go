package cache_test

import (
	b64 "encoding/base64"
	"errors"
	"soarca/internal/reporter/downstream_reporter/cache"
	"soarca/models/cacao"
	cache_model "soarca/models/cache"
	mock_time "soarca/test/unittest/mocks/mock_utils/time"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestReportWorkflowStartFirst(t *testing.T) {

	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time, 10)

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
		Description:   "test step",
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
		Name:                          "ssh-test-playbook",
		Description:                   "Playbook description",
		WorkflowStart:                 step1.ID,
		AuthenticationInfoDefinitions: map[string]cacao.AuthenticationInformation{"id": expectedAuth},
		AgentDefinitions:              map[string]cacao.AgentTarget{"agent1": expectedAgent},
		TargetDefinitions:             map[string]cacao.AgentTarget{"target1": expectedTarget},

		Workflow: map[string]cacao.Step{step1.ID: step1, end.ID: end},
	}
	executionId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	mock_time.On("Now").Return(timeNow)

	err = cacheReporter.ReportStepStart(executionId0, step1, cacao.NewVariables(expectedVariables), mock_time.Now())
	if err != nil {
		t.Fail()
	}

	mock_time.On("Now").Return(timeNow)
	err = cacheReporter.ReportStepEnd(executionId0, step1, cacao.NewVariables(), nil, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	expectedStarted, _ := time.Parse(layout, "2014-11-12T11:45:26.371Z")
	expectedEnded, _ := time.Parse(layout, "0001-01-01T00:00:00Z")
	expetedStepReport := cache_model.StepResult{
		ExecutionId: executionId0,
		StepId:      "action--test",
		Name:        "ssh-tests",
		Description: "test step",
		IsAutomated: true,
		Started:     timeNow,
		Ended:       timeNow,
		CommandsB64: []string{b64.StdEncoding.EncodeToString([]byte(expectedCommand.Command))},
		Variables:   cacao.NewVariables(),
		Status:      cache_model.SuccessfullyExecuted,
		Error:       nil,
	}

	expectedExecutions := []cache_model.ExecutionEntry{
		{
			ExecutionId: executionId0,
			PlaybookId:  "test",
			Name:        "ssh-test-playbook",
			Description: "Playbook description",
			Started:     expectedStarted,
			Ended:       expectedEnded,
			StepResults: map[string]cache_model.StepResult{expetedStepReport.StepId: expetedStepReport},
			Error:       nil,
			Status:      2,
		},
	}

	returnedExecutions, _ := cacheReporter.GetExecutions()

	exec, err := cacheReporter.GetExecutionReport(executionId0)
	assert.Equal(t, expectedExecutions, returnedExecutions)
	assert.Equal(t, len(expectedExecutions), 1)
	assert.Equal(t, expectedExecutions[0].ExecutionId, exec.ExecutionId)
	assert.Equal(t, expectedExecutions[0].PlaybookId, exec.PlaybookId)
	assert.Equal(t, expectedExecutions[0].StepResults, exec.StepResults)
	assert.Equal(t, expectedExecutions[0].Started, timeNow)
	assert.Equal(t, expectedExecutions[0].Ended, time.Time{})
	assert.Equal(t, expectedExecutions[0].Status, exec.Status)
	assert.Equal(t, err, nil)
	mock_time.AssertExpectations(t)
}

func TestReportWorkflowStartFifo(t *testing.T) {
	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time, 3)

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
		Description:   "step description",
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
		Name:                          "ssh-test-playbook",
		Description:                   "Playbook description",
		WorkflowStart:                 step1.ID,
		AuthenticationInfoDefinitions: map[string]cacao.AuthenticationInformation{"id": expectedAuth},
		AgentDefinitions:              map[string]cacao.AgentTarget{"agent1": expectedAgent},
		TargetDefinitions:             map[string]cacao.AgentTarget{"target1": expectedTarget},

		Workflow: map[string]cacao.Step{step1.ID: step1, end.ID: end},
	}
	executionId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	executionId1 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c1")
	executionId2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c2")
	executionId3 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c3")

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	executionIds := []uuid.UUID{
		executionId0,
		executionId1,
		executionId2,
		executionId3,
	}

	expectedStarted, _ := time.Parse(layout, "2014-11-12T11:45:26.371Z")
	expectedEnded, _ := time.Parse(layout, "0001-01-01T00:00:00Z")
	expectedExecutionsFull := []cache_model.ExecutionEntry{}
	for _, executionId := range executionIds[:len(executionIds)-1] {
		t.Log(executionId)
		entry := cache_model.ExecutionEntry{
			ExecutionId: executionId,
			PlaybookId:  "test",
			Name:        "ssh-test-playbook",
			Description: "Playbook description",
			Started:     expectedStarted,
			Ended:       expectedEnded,
			StepResults: map[string]cache_model.StepResult{},
			Error:       nil,
			Status:      2,
		}
		expectedExecutionsFull = append(expectedExecutionsFull, entry)
	}
	t.Log("")
	expectedExecutionsFifo := []cache_model.ExecutionEntry{}
	for _, executionId := range executionIds[1:] {
		t.Log(executionId)
		entry := cache_model.ExecutionEntry{
			ExecutionId: executionId,
			PlaybookId:  "test",
			Name:        "ssh-test-playbook",
			Description: "Playbook description",
			Started:     expectedStarted,
			Ended:       expectedEnded,
			StepResults: map[string]cache_model.StepResult{},
			Error:       nil,
			Status:      2,
		}
		expectedExecutionsFifo = append(expectedExecutionsFifo, entry)
	}

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId1, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId2, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	returnedExecutionsFull, _ := cacheReporter.GetExecutions()
	t.Log("expected")
	t.Log(expectedExecutionsFull)
	t.Log("returned")
	t.Log(returnedExecutionsFull)
	assert.Equal(t, expectedExecutionsFull, returnedExecutionsFull)

	err = cacheReporter.ReportWorkflowStart(executionId3, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	returnedExecutionsFifo, _ := cacheReporter.GetExecutions()
	assert.Equal(t, expectedExecutionsFifo, returnedExecutionsFifo)
	mock_time.AssertExpectations(t)
}

func TestReportWorkflowEnd(t *testing.T) {

	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time, 10)

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
		Description:   "step 1",
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
		Name:                          "ssh-test-playbook",
		Description:                   "Playbook description",
		WorkflowStart:                 step1.ID,
		AuthenticationInfoDefinitions: map[string]cacao.AuthenticationInformation{"id": expectedAuth},
		AgentDefinitions:              map[string]cacao.AgentTarget{"agent1": expectedAgent},
		TargetDefinitions:             map[string]cacao.AgentTarget{"target1": expectedTarget},

		Workflow: map[string]cacao.Step{step1.ID: step1, end.ID: end},
	}
	executionId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowEnd(executionId0, playbook, nil, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	expectedExecutionEntry := cache_model.ExecutionEntry{
		ExecutionId: executionId0,
		PlaybookId:  "test",
		Name:        "ssh-test-playbook",
		Description: "Playbook description",
		Started:     timeNow,
		Ended:       timeNow,
		StepResults: map[string]cache_model.StepResult{},
		Status:      cache_model.SuccessfullyExecuted,
	}
	expectedExecutions := []cache_model.ExecutionEntry{expectedExecutionEntry}

	returnedExecutions, _ := cacheReporter.GetExecutions()

	exec, err := cacheReporter.GetExecutionReport(executionId0)
	assert.Equal(t, expectedExecutions, returnedExecutions)
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
	cacheReporter := cache.New(mock_time, 10)

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
	executionId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepStart(executionId0, step1, cacao.NewVariables(expectedVariables), mock_time.Now())
	if err != nil {
		t.Fail()
	}

	expectedStepStatus := cache_model.StepResult{
		ExecutionId: executionId0,
		StepId:      step1.ID,
		Started:     timeNow,
		Ended:       time.Time{},
		Variables:   cacao.NewVariables(expectedVariables),
		Status:      cache_model.Ongoing,
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

	err = cacheReporter.ReportStepEnd(executionId0, step1, cacao.NewVariables(expectedVariables), nil, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	expectedStepResult := cache_model.StepResult{
		ExecutionId: executionId0,
		StepId:      step1.ID,
		Started:     timeNow,
		Ended:       timeNow,
		Variables:   cacao.NewVariables(expectedVariables),
		Status:      cache_model.SuccessfullyExecuted,
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

func TestReportStepStartCommandsEncoding(t *testing.T) {
	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time, 10)

	expectedCommand1 := cacao.Command{
		Type:       "manual",
		CommandB64: b64.StdEncoding.EncodeToString([]byte("do ssh ls -la in the terminal")),
	}
	expectedCommand2 := cacao.Command{
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
		Commands:      []cacao.Command{expectedCommand1, expectedCommand2},
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
	executionId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepStart(executionId0, step1, cacao.NewVariables(expectedVariables), mock_time.Now())
	if err != nil {
		t.Fail()
	}

	encodedCommand1 := expectedCommand1.CommandB64
	encodedCommand2 := b64.StdEncoding.EncodeToString([]byte(expectedCommand2.Command))
	expectedCommandsB64 := []string{encodedCommand1, encodedCommand2}

	expectedStepStatus := cache_model.StepResult{
		ExecutionId: executionId0,
		StepId:      step1.ID,
		Started:     timeNow,
		Ended:       time.Time{},
		Variables:   cacao.NewVariables(expectedVariables),
		Status:      cache_model.Ongoing,
		CommandsB64: expectedCommandsB64,
		Error:       nil,
		IsAutomated: false,
	}

	exec, err := cacheReporter.GetExecutionReport(executionId0)
	stepStatus := exec.StepResults[step1.ID]
	t.Log("stepStatus commands")
	t.Log(stepStatus.CommandsB64)
	t.Log("expectedStep commands")
	t.Log(expectedStepStatus.CommandsB64)
	assert.Equal(t, stepStatus.ExecutionId, expectedStepStatus.ExecutionId)
	assert.Equal(t, stepStatus.StepId, expectedStepStatus.StepId)
	assert.Equal(t, stepStatus.Started, expectedStepStatus.Started)
	assert.Equal(t, stepStatus.Ended, expectedStepStatus.Ended)
	assert.Equal(t, stepStatus.Variables, expectedStepStatus.Variables)
	assert.Equal(t, stepStatus.Status, expectedStepStatus.Status)
	assert.Equal(t, stepStatus.Error, expectedStepStatus.Error)
	assert.Equal(t, stepStatus.CommandsB64, expectedStepStatus.CommandsB64)
	assert.Equal(t, stepStatus.IsAutomated, expectedStepStatus.IsAutomated)
	assert.Equal(t, err, nil)

	err = cacheReporter.ReportStepEnd(executionId0, step1, cacao.NewVariables(expectedVariables), nil, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	expectedStepResult := cache_model.StepResult{
		ExecutionId: executionId0,
		StepId:      step1.ID,
		Started:     timeNow,
		Ended:       timeNow,
		Variables:   cacao.NewVariables(expectedVariables),
		Status:      cache_model.SuccessfullyExecuted,
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

func TestReportStepStartManualCommand(t *testing.T) {
	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time, 10)

	expectedCommand := cacao.Command{
		Type:    "manual",
		Command: "do ssh ls -la in the terminal",
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
	executionId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepStart(executionId0, step1, cacao.NewVariables(expectedVariables), mock_time.Now())
	if err != nil {
		t.Fail()
	}

	encodedCommand := b64.StdEncoding.EncodeToString([]byte(expectedCommand.Command))

	expectedStepStatus := cache_model.StepResult{
		ExecutionId: executionId0,
		StepId:      step1.ID,
		Started:     timeNow,
		Ended:       time.Time{},
		Variables:   cacao.NewVariables(expectedVariables),
		Status:      cache_model.Ongoing,
		CommandsB64: []string{encodedCommand},
		Error:       nil,
		IsAutomated: false,
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
	assert.Equal(t, stepStatus.CommandsB64, expectedStepStatus.CommandsB64)
	assert.Equal(t, stepStatus.IsAutomated, expectedStepStatus.IsAutomated)
	assert.Equal(t, err, nil)

	err = cacheReporter.ReportStepEnd(executionId0, step1, cacao.NewVariables(expectedVariables), nil, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	expectedStepResult := cache_model.StepResult{
		ExecutionId: executionId0,
		StepId:      step1.ID,
		Started:     timeNow,
		Ended:       timeNow,
		Variables:   cacao.NewVariables(expectedVariables),
		Status:      cache_model.SuccessfullyExecuted,
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

func TestInvalidStepReportAfterStepEnd(t *testing.T) {
	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time, 10)

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
	executionId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepStart(executionId0, step1, cacao.NewVariables(expectedVariables), mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepEnd(executionId0, step1, cacao.NewVariables(expectedVariables), nil, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	err = cacheReporter.ReportStepEnd(executionId0, step1, cacao.NewVariables(expectedVariables), nil, mock_time.Now())
	if err == nil {
		t.Fail()
	}

	expectedErr := errors.New("step status precondition not met for step update [step status: successfully_executed]")
	assert.Equal(t, err, expectedErr)
	mock_time.AssertExpectations(t)
}

func TestAcceptedStepReportAfterExecutionEnd(t *testing.T) {

	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time, 10)

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
	executionId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepStart(executionId0, step1, cacao.NewVariables(expectedVariables), mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowEnd(executionId0, playbook, nil, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	err = cacheReporter.ReportStepEnd(executionId0, step1, cacao.NewVariables(expectedVariables), nil, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	mock_time.AssertExpectations(t)
}
