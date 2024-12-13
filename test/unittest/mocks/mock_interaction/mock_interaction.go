package mock_interaction

import (
	"soarca/pkg/models/manual"

	"github.com/stretchr/testify/mock"
)

type MockInteraction struct {
	mock.Mock
}

func (mock *MockInteraction) Queue(command manual.InteractionCommand,
	channel chan manual.InteractionResponse) error {
	args := mock.Called(command, channel)
	return args.Error(0)
}
