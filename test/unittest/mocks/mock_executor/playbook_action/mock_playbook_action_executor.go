package mock_playbook_action_executor

import (
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"

	"github.com/stretchr/testify/mock"
)

type Mock_PlaybookActionExecutor struct {
	mock.Mock
}

func (executor *Mock_PlaybookActionExecutor) Execute(metadata execution.Metadata,
	step cacao.Step,
	variables cacao.Variables) (cacao.Variables, error) {
	args := executor.Called(metadata, step, variables)
	return args.Get(0).(cacao.Variables), args.Error(1)
}
