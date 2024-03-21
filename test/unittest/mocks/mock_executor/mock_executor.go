package mock_executor

import (
	"soarca/internal/executors/action"
	"soarca/models/cacao"
	"soarca/models/execution"

	"github.com/stretchr/testify/mock"
)

type Mock_Action_Executor struct {
	mock.Mock
}

func (executer *Mock_Action_Executor) Execute(
	metadata execution.Metadata,
	details action.StepDetails) (cacao.Variables,
	error) {
	args := executer.Called(metadata, details)
	return args.Get(0).(cacao.Variables), args.Error(1)
}
