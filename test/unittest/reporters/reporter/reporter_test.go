package reporter_test

import (
	"errors"
	"soarca/internal/reporter"
	ds_reporter "soarca/internal/reporter/downstream_reporter"
	"soarca/models/cacao"
	"soarca/test/unittest/mocks/mock_reporter"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

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
	if err == nil {
		t.Fail()
	}
	expected_err := errors.New("attempting to register too many reporters")
	assert.Equal(t, expected_err, err)
}

func TestReportWorkflow(t *testing.T) {
	mock_ds_reporter := mock_reporter.Mock_Downstream_Reporter{}
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

	mock_ds_reporter.On("ReportWorkflow", executionId, playbook).Return(nil)
	reporter.ReportWorkflow(executionId, playbook)
}

func TestReportStep(t *testing.T) {
	mock_ds_reporter := mock_reporter.Mock_Downstream_Reporter{}
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

	mock_ds_reporter.On("ReportStep", executionId, step1, cacao.NewVariables(expectedVariables), nil).Return(nil)
	reporter.ReportStep(executionId, step1, cacao.NewVariables(expectedVariables), nil)
}
