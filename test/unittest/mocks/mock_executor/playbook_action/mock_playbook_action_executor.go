package mock_playbook_action_executor

import (
	"soarca/models/cacao"
	"soarca/models/execution"

	"github.com/stretchr/testify/mock"
)

type Mock_Playybook_Action_Executor struct {
	mock.Mock
}

func (executer *Mock_Playybook_Action_Executor) Execute(metadata execution.Metadata,
	step cacao.Step,
	variables cacao.Variables) (cacao.Variables, error) {
	args := executer.Called(metadata, step, variables)
	return args.Get(0).(cacao.Variables), args.Error(1)
}
