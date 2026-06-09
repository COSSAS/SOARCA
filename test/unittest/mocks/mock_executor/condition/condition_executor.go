package mock_condition_executor

import (
	"soarca/pkg/core/executors"
	"soarca/pkg/models/execution"

	"github.com/stretchr/testify/mock"
)

type Mock_Condition struct {
	mock.Mock
}

func (executor *Mock_Condition) Execute(metadata execution.Metadata,
	context executors.Context) (string, bool, error) {
	args := executor.Called(metadata, context)
	return args.String(0), args.Bool(1), args.Error(2)
}
