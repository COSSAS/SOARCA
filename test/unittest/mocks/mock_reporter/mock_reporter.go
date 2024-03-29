package mock_reporter

import (
	"soarca/models/cacao"

	"github.com/stretchr/testify/mock"
)

type Mock_Reporter struct {
	mock.Mock
}

func (decomposer_reporter *Mock_Reporter) ReportWorkflow(workflow cacao.Workflow) (interface{}, error) {
	return new(interface{}), nil
}

func (decomposer_reporter *Mock_Reporter) ReportStep(step cacao.Step, vars cacao.Variables, err error) (interface{}, error) {
	return new(interface{}), nil
}
