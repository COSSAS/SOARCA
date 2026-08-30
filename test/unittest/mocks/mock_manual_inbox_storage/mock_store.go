package mock_manual_inbox_storage

import (
	"soarca/internal/manual/model"
	"soarca/internal/runs/model"

	"github.com/stretchr/testify/mock"
)

type MockInboxStorage struct {
	mock.Mock
}

func (mock *MockInboxStorage) GetPendingCommands() ([]manual.CommandInfo, error) {
	args := mock.Called()
	return args.Get(0).([]manual.CommandInfo), args.Error(1)
}

func (mock *MockInboxStorage) ListPendingCommands() ([]manual.CommandInfo, error) {
	return mock.GetPendingCommands()
}

func (mock *MockInboxStorage) GetPendingCommand(metadata run.Metadata) (manual.CommandInfo, error) {
	args := mock.Called(metadata)
	return args.Get(0).(manual.CommandInfo), args.Error(1)
}

func (mock *MockInboxStorage) PostContinue(response manual.Response) error {
	args := mock.Called(response)
	return args.Error(0)
}

func (mock *MockInboxStorage) ContinuePendingCommand(response manual.Response) error {
	return mock.PostContinue(response)
}
