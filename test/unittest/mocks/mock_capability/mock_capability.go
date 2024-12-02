package mock_capability

import (
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"

	"github.com/stretchr/testify/mock"
)

type Mock_Capability struct {
	mock.Mock
}

func (capability *Mock_Capability) Execute(metadata execution.Metadata,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variables cacao.Variables) (cacao.Variables, error) {
	args := capability.Called(metadata, command, authentication, target, variables)
	return args.Get(0).(cacao.Variables), args.Error(1)
}

func (capability *Mock_Capability) GetType() string {
	args := capability.Called()
	return args.Get(0).(string)
}
