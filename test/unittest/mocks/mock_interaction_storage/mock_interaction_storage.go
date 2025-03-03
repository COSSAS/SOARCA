package mock_interaction_storage

import (
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"

	"github.com/stretchr/testify/mock"
)

type MockInteractionStorage struct {
	mock.Mock
}

func (mock *MockInteractionStorage) GetPendingCommands() ([]manual.CommandInfo, error) {
	args := mock.Called()
	return args.Get(0).([]manual.CommandInfo), args.Error(1)
}

func (mock *MockInteractionStorage) GetPendingCommand(metadata execution.Metadata) (manual.CommandInfo, error) {
	args := mock.Called(metadata)
	return args.Get(0).(manual.CommandInfo), args.Error(1)
}

func (mock *MockInteractionStorage) PostContinue(response manual.InteractionResponse) error {
	args := mock.Called(response)
	return args.Error(0)
}
