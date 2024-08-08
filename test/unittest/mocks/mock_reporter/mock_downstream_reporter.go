package mock_reporter

import (
	"soarca/models/cacao"
	"sync"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Downstream_Reporter struct {
	mock.Mock
	Wg *sync.WaitGroup
}

func (ds_reporter *Mock_Downstream_Reporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) error {
	defer ds_reporter.Wg.Done()
	args := ds_reporter.Called(executionId, playbook)
	return args.Error(0)
}
func (ds_reporter *Mock_Downstream_Reporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error) error {
	defer ds_reporter.Wg.Done()
	args := ds_reporter.Called(executionId, playbook, workflowError)
	return args.Error(0)
}

func (ds_reporter *Mock_Downstream_Reporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables) error {
	defer ds_reporter.Wg.Done()
	args := ds_reporter.Called(executionId, step, stepResults)
	return args.Error(0)
}
func (ds_reporter *Mock_Downstream_Reporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, stepError error) error {
	defer ds_reporter.Wg.Done()
	args := ds_reporter.Called(executionId, step, stepResults, stepError)
	return args.Error(0)
}
