package mock_reporter

import (
	"soarca/models/cacao"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Reporter struct {
	mock.Mock
}

func (reporter *Mock_Reporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) {
	_ = reporter.Called(executionId, playbook)
}
func (reporter *Mock_Reporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook) {
	_ = reporter.Called(executionId, playbook)
}

func (reporter *Mock_Reporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, err error) {
	_ = reporter.Called(executionId, step, returnVars, err)
}
func (reporter *Mock_Reporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, err error) {
	_ = reporter.Called(executionId, step, returnVars, err)
}
