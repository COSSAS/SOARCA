package mock_cache

import (
	cache_model "soarca/models/cache"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Cache struct {
	mock.Mock
}

func (reporter *Mock_Cache) GetExecutions() ([]cache_model.ExecutionEntry, error) {
	args := reporter.Called()
	return args.Get(0).([]cache_model.ExecutionEntry), args.Error(1)
}

func (reporter *Mock_Cache) GetExecutionReport(executionKey uuid.UUID) (cache_model.ExecutionEntry, error) {
	args := reporter.Called(executionKey)
	return args.Get(0).(cache_model.ExecutionEntry), args.Error(1)
}
