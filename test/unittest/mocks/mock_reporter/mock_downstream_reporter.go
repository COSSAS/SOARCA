package mock_reporter

import (
	"soarca/models/cacao"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Downstream_Reporter struct {
	mock.Mock
}

func (reporter *Mock_Downstream_Reporter) ReportWorkflow(executionId uuid.UUID, playbook cacao.Playbook) error {
	args := reporter.Called(executionId, playbook)
	return args.Error(0)
}

func (reporter *Mock_Downstream_Reporter) ReportStep(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error) error {
	args := reporter.Called(executionId, step, stepResults, err)
	return args.Error(0)
}
