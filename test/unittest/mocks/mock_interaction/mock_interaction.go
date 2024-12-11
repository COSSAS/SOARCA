package mock_interaction

import (
	"soarca/pkg/interaction"

	"github.com/stretchr/testify/mock"
)

type MockInteraction struct {
	mock.Mock
}

func (mock *MockInteraction) Queue(command interaction.InteractionCommand,
	channel chan interaction.InteractionResponse) error {
	args := mock.Called(command, channel)
	return args.Error(0)
}
