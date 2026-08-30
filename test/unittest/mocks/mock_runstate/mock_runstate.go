package mock_runstate

import (
	runstate_model "soarca/internal/runs/state"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockRunState struct {
	mock.Mock
}

func (reporter *MockRunState) GetRuns() ([]runstate_model.RunEntry, error) {
	args := reporter.Called()
	return args.Get(0).([]runstate_model.RunEntry), args.Error(1)
}

func (reporter *MockRunState) GetRunReport(runKey uuid.UUID) (runstate_model.RunEntry, error) {
	args := reporter.Called(runKey)
	return args.Get(0).(runstate_model.RunEntry), args.Error(1)
}
