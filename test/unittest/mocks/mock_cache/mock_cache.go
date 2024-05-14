package mock_cache

import (
	"soarca/models/cacao"

	cache_model "soarca/models/cache"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Cache struct {
	mock.Mock
}

func (reporter *Mock_Cache) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) error {
	args := reporter.Called(executionId, playbook)
	return args.Error(0)
}
func (reporter *Mock_Cache) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error) error {
	args := reporter.Called(executionId, playbook, workflowError)
	return args.Error(0)
}

func (reporter *Mock_Cache) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables) error {
	args := reporter.Called(executionId, step, stepResults)
	return args.Error(0)
}
func (reporter *Mock_Cache) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, stepError error) error {
	args := reporter.Called(executionId, step, stepResults, stepError)
	return args.Error(0)
}

func (reporter *Mock_Cache) GetExecutions() ([]cache_model.ExecutionEntry, error) {
	args := reporter.Called()
	return args.Get(0).([]cache_model.ExecutionEntry), args.Error(1)
}

func (reporter *Mock_Cache) GetExecutionReport(executionKey uuid.UUID) (cache_model.ExecutionEntry, error) {
	args := reporter.Called(executionKey)
	return args.Get(0).(cache_model.ExecutionEntry), args.Error(1)
}
