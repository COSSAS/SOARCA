package mock_executor

import (
	"soarca/internal/executer"
	"soarca/models/cacao"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Executor struct {
	mock.Mock
	OnCompletionCallback func(executionId uuid.UUID, output map[string]cacao.Variables)
}

func (executer *Mock_Executor) ExecuteAsync(command cacao.Command,
	variable map[string]cacao.Variables,
	module string,
	callback func(executionId uuid.UUID,
		output map[string]cacao.Variables)) error {

	executer.OnCompletionCallback = callback
	args := executer.Called(command, variable, module, callback)
	return args.Error(0)
}

func (executer *Mock_Executor) Execute(
	executionId uuid.UUID,
	command cacao.Command,
	variable map[string]cacao.Variables,
	module string) (uuid.UUID,
	map[string]cacao.Variables,
	error) {
	args := executer.Called(executionId, command, variable, module)
	return args.Get(0).(uuid.UUID), args.Get(1).(map[string]cacao.Variables), args.Error(2)
}

func (executer *Mock_Executor) Pause(
	command executer.CommandData,
	module string) error {

	args := executer.Called()
	return args.Error(0)
}

func (executer *Mock_Executor) Resume(
	command executer.CommandData,
	module string) error {

	args := executer.Called()
	return args.Error(0)
}

func (executer *Mock_Executor) Kill(command executer.CommandData, module string) error {
	args := executer.Called()
	return args.Error(0)
}
