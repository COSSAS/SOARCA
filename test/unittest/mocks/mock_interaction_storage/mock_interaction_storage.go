package mock_interaction_storage

import (
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"

	"github.com/stretchr/testify/mock"
)

type MockInteractionStorage struct {
	mock.Mock
}

func (mock *MockInteractionStorage) GetPendingCommands() ([]manual.InteractionCommandData, int, error) {
	args := mock.Called()
	return args.Get(0).([]manual.InteractionCommandData), args.Int(1), args.Error(2)
}

func (mock *MockInteractionStorage) GetPendingCommand(metadata execution.Metadata) (manual.InteractionCommandData, int, error) {
	args := mock.Called(metadata)
	return args.Get(0).(manual.InteractionCommandData), args.Int(1), args.Error(2)
}

func (mock *MockInteractionStorage) PostContinue(outArgsResult manual.ManualOutArgsUpdatePayload) (int, error) {
	args := mock.Called(outArgsResult)
	return args.Int(0), args.Error(1)
}
