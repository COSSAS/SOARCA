package mock_capability

import (
	"soarca/models/cacao"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Capability struct {
	mock.Mock
}

func (capability *Mock_Capability) Execute(executionId uuid.UUID,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.Target,
	variables map[string]cacao.Variables) (map[string]cacao.Variables, error) {
	args := capability.Called(executionId, command, authentication, target, variables)
	return args.Get(0).(map[string]cacao.Variables), args.Error(1)
}

func (capability *Mock_Capability) GetType() string {
	args := capability.Called()
	return args.Get(0).(string)
}
