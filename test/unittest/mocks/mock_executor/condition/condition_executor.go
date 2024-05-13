package mock_condition_executor

import (
	"soarca/models/cacao"
	"soarca/models/execution"

	"github.com/stretchr/testify/mock"
)

type Mock_Condition struct {
	mock.Mock
}

func (executer *Mock_Condition) Execute(metadata execution.Metadata,
	step cacao.Step,
	variables cacao.Variables) (string, bool, error) {
	args := executer.Called(metadata, step, variables)
	return args.String(0), args.Bool(1), args.Error(2)
}
