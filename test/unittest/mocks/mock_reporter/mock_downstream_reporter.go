package mock_reporter

import (
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Downstream_Reporter struct {
	mock.Mock
	Wg *sync.WaitGroup
}

func (ds_reporter *Mock_Downstream_Reporter) ReportWorkflowStart(runId uuid.UUID, playbook cacao.Playbook, at time.Time) error {
	defer ds_reporter.Wg.Done()
	args := ds_reporter.Called(runId, playbook, at)
	return args.Error(0)
}
func (ds_reporter *Mock_Downstream_Reporter) ReportWorkflowEnd(runId uuid.UUID, playbook cacao.Playbook, workflowError error, at time.Time) error {
	defer ds_reporter.Wg.Done()
	args := ds_reporter.Called(runId, playbook, workflowError, at)
	return args.Error(0)
}

func (ds_reporter *Mock_Downstream_Reporter) ReportStepStart(metadata run.Metadata, step cacao.Step, stepResults cacao.Variables, at time.Time) error {
	defer ds_reporter.Wg.Done()
	args := ds_reporter.Called(metadata, step, stepResults, at)
	return args.Error(0)
}
func (ds_reporter *Mock_Downstream_Reporter) ReportStepEnd(metadata run.Metadata, step cacao.Step, stepResults cacao.Variables, stepError error, at time.Time) error {
	defer ds_reporter.Wg.Done()
	args := ds_reporter.Called(metadata, step, stepResults, stepError, at)
	return args.Error(0)
}
