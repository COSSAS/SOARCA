package mock_reporter

import (
	"soarca/models/cacao"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Downstream_Reporter struct {
	mock.Mock
	Wg *sync.WaitGroup
}

func (ds_reporter *Mock_Downstream_Reporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook, at time.Time) error {
	defer ds_reporter.Wg.Done()
	args := ds_reporter.Called(executionId, playbook, at)
	return args.Error(0)
}
func (ds_reporter *Mock_Downstream_Reporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error, at time.Time) error {
	defer ds_reporter.Wg.Done()
	args := ds_reporter.Called(executionId, playbook, workflowError, at)
	return args.Error(0)
}

func (ds_reporter *Mock_Downstream_Reporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, at time.Time) error {
	defer ds_reporter.Wg.Done()
	args := ds_reporter.Called(executionId, step, stepResults, at)
	return args.Error(0)
}
func (ds_reporter *Mock_Downstream_Reporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, stepError error, at time.Time) error {
	defer ds_reporter.Wg.Done()
	args := ds_reporter.Called(executionId, step, stepResults, stepError, at)
	return args.Error(0)
}
