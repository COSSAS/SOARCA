package mock_executor

import (
	"soarca/internal/workflow/steps"
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"

	"github.com/stretchr/testify/mock"
)

type Mock_Action_Executor struct {
	mock.Mock
}

func (executer *Mock_Action_Executor) Execute(
	metadata run.Metadata,
	details executors.PlaybookStepMetadata) (cacao.Variables,
	error) {
	args := executer.Called(metadata, details)
	return args.Get(0).(cacao.Variables), args.Error(1)
}
