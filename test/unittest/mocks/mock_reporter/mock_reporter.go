package mock_reporter

import (
	"soarca/models/cacao"
	"soarca/models/execution"

	"github.com/stretchr/testify/mock"
)

type Mock_Reporter struct {
	mock.Mock
}

func (reporter *Mock_Reporter) ReportWorkflow(executionContext execution.Metadata, playbook cacao.Playbook) {

}

func (reporter *Mock_Reporter) ReportStep(executionContext execution.Metadata, step cacao.Step, outVars cacao.Variables, err error) {

}
