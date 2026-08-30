package mock_condition_executor

import (
	"soarca/internal/workflow/steps"
	"soarca/internal/runs/model"

	"github.com/stretchr/testify/mock"
)

type Mock_Condition struct {
	mock.Mock
}

func (executer *Mock_Condition) Execute(metadata run.Metadata,
	context executors.Context) (string, bool, error) {
	args := executer.Called(metadata, context)
	return args.String(0), args.Bool(1), args.Error(2)
}
