package mock_reporter

import (
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Reporter struct {
	mock.Mock
}

func (reporter *Mock_Reporter) ReportWorkflowStart(runId uuid.UUID, playbook cacao.Playbook, at time.Time) {
	_ = reporter.Called(runId, playbook, at)
}
func (reporter *Mock_Reporter) ReportWorkflowEnd(runId uuid.UUID, playbook cacao.Playbook, err error, at time.Time) {
	_ = reporter.Called(runId, playbook, err, at)
}

func (reporter *Mock_Reporter) ReportStepStart(metadata run.Metadata, step cacao.Step, returnVars cacao.Variables, at time.Time) {
	_ = reporter.Called(metadata, step, returnVars, at)
}
func (reporter *Mock_Reporter) ReportStepEnd(metadata run.Metadata, step cacao.Step, returnVars cacao.Variables, err error, at time.Time) {
	_ = reporter.Called(metadata, step, returnVars, err, at)
}
