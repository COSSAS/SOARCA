package mock_executor

import (
	"soarca/models/cacao"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Executor struct {
	mock.Mock
	OnCompletionCallback func(executionId uuid.UUID, output map[string]cacao.Variables)
}

func (executer *Mock_Executor) Execute(
	executionId uuid.UUID,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variable map[string]cacao.Variables,
	agent cacao.AgentTarget) (uuid.UUID,
	map[string]cacao.Variables,
	error) {
	args := executer.Called(executionId, command, authentication, target, variable, agent)
	return args.Get(0).(uuid.UUID), args.Get(1).(map[string]cacao.Variables), args.Error(2)
}
