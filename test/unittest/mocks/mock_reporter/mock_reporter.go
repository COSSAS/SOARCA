package mock_reporter

import (
	"soarca/pkg/models/cacao"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Reporter struct {
	mock.Mock
}

func (reporter *Mock_Reporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook, at time.Time) {
	_ = reporter.Called(executionId, playbook, at)
}
func (reporter *Mock_Reporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, err error, at time.Time) {
	_ = reporter.Called(executionId, playbook, err, at)
}

func (reporter *Mock_Reporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, at time.Time) {
	_ = reporter.Called(executionId, step, returnVars, at)
}
func (reporter *Mock_Reporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, err error, at time.Time) {
	_ = reporter.Called(executionId, step, returnVars, err, at)
}
