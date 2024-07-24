package mock_capability

import (
	"soarca/models/cacao"
	"soarca/models/execution"

	"github.com/stretchr/testify/mock"
)

type Mock_Capability struct {
	mock.Mock
}

func (capability *Mock_Capability) Execute(metadata execution.Metadata,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	inputVariables cacao.Variables,
	outputVariables cacao.Variables) (cacao.Variables, error) {
	args := capability.Called(metadata,
		command,
		authentication,
		target,
		inputVariables,
		outputVariables)
	return args.Get(0).(cacao.Variables), args.Error(1)
}

func (capability *Mock_Capability) GetType() string {
	args := capability.Called()
	return args.Get(0).(string)
}
