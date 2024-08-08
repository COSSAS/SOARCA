package reporter_test

import (
	"errors"
	"soarca/internal/reporter"
	ds_reporter "soarca/internal/reporter/downstream_reporter"
	"soarca/models/cacao"
	"soarca/test/unittest/mocks/mock_reporter"
	"sync"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

// NOTE: reporter functions call go routines with loops inside, which seems to break testing expectations
// This SO post seems to propose a valid solution: https://stackoverflow.com/questions/51065482/how-to-test-if-a-goroutine-has-been-called-while-unit-testing-in-golang

func TestRegisterReporter(t *testing.T) {
	mock_ds_reporter := mock_reporter.Mock_Downstream_Reporter{}
	reporter := reporter.New([]ds_reporter.IDownStreamReporter{})
	err := reporter.RegisterReporters([]ds_reporter.IDownStreamReporter{&mock_ds_reporter})
	if err != nil {
		t.Fail()
	}
}

func TestRegisterTooManyReporters(t *testing.T) {
	too_many_reporters := make([]ds_reporter.IDownStreamReporter, reporter.MaxReporters+1)
	mock_ds_reporter := mock_reporter.Mock_Downstream_Reporter{}
	for i := range too_many_reporters {
		too_many_reporters[i] = &mock_ds_reporter
	}

	reporter := reporter.New([]ds_reporter.IDownStreamReporter{})
	err := reporter.RegisterReporters(too_many_reporters)

	expected_err := errors.New("attempting to register too many reporters")
	assert.Equal(t, expected_err, err)
}

func TestReportWorkflowStart(t *testing.T) {
	var wg sync.WaitGroup
	mock_ds_reporter := mock_reporter.Mock_Downstream_Reporter{Wg: &wg}
	reporter := reporter.New([]ds_reporter.IDownStreamReporter{&mock_ds_reporter})

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

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	wg.Add(1)
	mock_ds_reporter.On("ReportWorkflowStart", executionId, playbook).Return(nil)
	reporter.ReportWorkflowStart(executionId, playbook)

	wg.Wait()
	mock_ds_reporter.AssertExpectations(t)
}

func TestReportWorkflowEnd(t *testing.T) {
	var wg sync.WaitGroup
	mock_ds_reporter := mock_reporter.Mock_Downstream_Reporter{Wg: &wg}
	reporter := reporter.New([]ds_reporter.IDownStreamReporter{&mock_ds_reporter})

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

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	wg.Add(1)
	mock_ds_reporter.On("ReportWorkflowEnd", executionId, playbook, nil).Return(nil)
	reporter.ReportWorkflowEnd(executionId, playbook, nil)

	wg.Wait()
	mock_ds_reporter.AssertExpectations(t)
}

func TestReportStepStart(t *testing.T) {
	var wg sync.WaitGroup
	mock_ds_reporter := mock_reporter.Mock_Downstream_Reporter{Wg: &wg}
	reporter := reporter.New([]ds_reporter.IDownStreamReporter{&mock_ds_reporter})

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

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	wg.Add(1)
	mock_ds_reporter.On("ReportStepStart", executionId, step1, cacao.NewVariables(expectedVariables)).Return(nil)
	reporter.ReportStepStart(executionId, step1, cacao.NewVariables(expectedVariables))

	wg.Wait()
	mock_ds_reporter.AssertExpectations(t)
}

func TestReportStepEnd(t *testing.T) {
	var wg sync.WaitGroup
	mock_ds_reporter := mock_reporter.Mock_Downstream_Reporter{Wg: &wg}
	reporter := reporter.New([]ds_reporter.IDownStreamReporter{&mock_ds_reporter})

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

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	wg.Add(1)
	mock_ds_reporter.On("ReportStepEnd", executionId, step1, cacao.NewVariables(expectedVariables), nil).Return(nil)
	reporter.ReportStepEnd(executionId, step1, cacao.NewVariables(expectedVariables), nil)

	wg.Wait()
	mock_ds_reporter.AssertExpectations(t)
}

func TestMultipleDownstreamReporters(t *testing.T) {
	var wg sync.WaitGroup
	mock_ds_reporter1 := mock_reporter.Mock_Downstream_Reporter{Wg: &wg}
	mock_ds_reporter2 := mock_reporter.Mock_Downstream_Reporter{Wg: &wg}
	reporter := reporter.New([]ds_reporter.IDownStreamReporter{&mock_ds_reporter1, &mock_ds_reporter2})
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

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	wg.Add(2)
	mock_ds_reporter1.On("ReportWorkflowStart", executionId, playbook).Return(nil)
	mock_ds_reporter2.On("ReportWorkflowStart", executionId, playbook).Return(nil)

	wg.Add(2)
	mock_ds_reporter1.On("ReportStepStart", executionId, step1, cacao.NewVariables(expectedVariables)).Return(nil)
	mock_ds_reporter2.On("ReportStepStart", executionId, step1, cacao.NewVariables(expectedVariables)).Return(nil)

	wg.Add(2)
	mock_ds_reporter1.On("ReportStepEnd", executionId, step1, cacao.NewVariables(expectedVariables), nil).Return(nil)
	mock_ds_reporter2.On("ReportStepEnd", executionId, step1, cacao.NewVariables(expectedVariables), nil).Return(nil)

	wg.Add(2)
	mock_ds_reporter1.On("ReportWorkflowEnd", executionId, playbook, nil).Return(nil)
	mock_ds_reporter2.On("ReportWorkflowEnd", executionId, playbook, nil).Return(nil)

	reporter.ReportWorkflowStart(executionId, playbook)
	reporter.ReportStepStart(executionId, step1, cacao.NewVariables(expectedVariables))
	reporter.ReportStepEnd(executionId, step1, cacao.NewVariables(expectedVariables), nil)
	reporter.ReportWorkflowEnd(executionId, playbook, nil)

	wg.Wait()
	mock_ds_reporter1.AssertExpectations(t)
	mock_ds_reporter2.AssertExpectations(t)
}
