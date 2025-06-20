package mock_executor

import (
	"soarca/pkg/core/executors"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"

	"github.com/stretchr/testify/mock"
)

type Mock_Action_Executor struct {
	mock.Mock
}

func (executor *Mock_Action_Executor) Execute(
	metadata execution.Metadata,
	details executors.PlaybookStepMetadata) (cacao.Variables,
	error) {
	args := executor.Called(metadata, details)
	return args.Get(0).(cacao.Variables), args.Error(1)
}
