package mock_reporter

import (
	"soarca/models/cacao"

	cache_model "soarca/models/cache"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Downstream_Reporter struct {
	mock.Mock
}

func (reporter *Mock_Downstream_Reporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) error {
	args := reporter.Called(executionId, playbook)
	return args.Error(0)
}
func (reporter *Mock_Downstream_Reporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error) error {
	args := reporter.Called(executionId, playbook, workflowError)
	return args.Error(0)
}

func (reporter *Mock_Downstream_Reporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables) error {
	args := reporter.Called(executionId, step, stepResults)
	return args.Error(0)
}
func (reporter *Mock_Downstream_Reporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, stepError error) error {
	args := reporter.Called(executionId, step, stepResults, stepError)
	return args.Error(0)
}

func (reporter *Mock_Downstream_Reporter) GetExecutions() ([]cache_model.ExecutionEntry, error) {
	args := reporter.Called()
	return args.Get(0).([]cache_model.ExecutionEntry), args.Error(1)
}

func (reporter *Mock_Downstream_Reporter) GetExecutionReport(executionKey uuid.UUID) (cache_model.ExecutionEntry, error) {
	args := reporter.Called(executionKey)
	return args.Get(0).(cache_model.ExecutionEntry), args.Error(1)
}
