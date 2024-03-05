package mock_executor

import (
	"soarca/models/cacao"
	"soarca/models/execution"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Executor struct {
	mock.Mock
	OnCompletionCallback func(executionId uuid.UUID, output cacao.Variables)
}

func (executer *Mock_Executor) Execute(
	metadata execution.Metadata,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variable cacao.Variables,
	agent cacao.AgentTarget) (uuid.UUID,
	cacao.Variables,
	error) {
	args := executer.Called(metadata, command, authentication, target, variable, agent)
	return args.Get(0).(uuid.UUID), args.Get(1).(cacao.Variables), args.Error(2)
}
