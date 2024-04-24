package mock_reporter

import (
	"soarca/models/cacao"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Reporter struct {
	mock.Mock
}

func (reporter *Mock_Reporter) ReportWorkflow(executionId uuid.UUID, playbook cacao.Playbook) {
	_ = reporter.Called(executionId, playbook)
}

func (reporter *Mock_Reporter) ReportStep(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, err error) {
	_ = reporter.Called(executionId, step, returnVars, err)
}
