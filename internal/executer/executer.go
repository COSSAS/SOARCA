package executer

import (
	"soarca/models/cacao"

	"github.com/google/uuid"
)

type CommandData struct {
	// id       string
	// command  string
	// isBase64 bool
}

type Variable struct{}
type AuthInfo struct{}
type Target struct{}
type Module struct{}

type IExecuter interface {
	Execute(executionId uuid.UUID,
		command cacao.Command,
		variable map[string]cacao.Variables,
		module string) (uuid.UUID, map[string]cacao.Variables, error)
	ExecuteAsync(command cacao.Command, variable map[string]cacao.Variables, module string, callback func(uuid.UUID, map[string]cacao.Variables)) error
	Pause(command CommandData, module string) error
	Resume(command CommandData, module string) error
	Kill(command CommandData, module string) error
}

type Executer struct {
}

func (executer *Executer) Execute(executionId uuid.UUID,
	command cacao.Command,
	variable map[string]cacao.Variables,
	module string) (uuid.UUID, map[string]cacao.Variables, error) {

	var vars = map[string]cacao.Variables{"test": {Name: "test"}}
	var id = uuid.New()

	return id, vars, nil
}

func (executer *Executer) ExecuteAsync(command cacao.Command,
	variable map[string]cacao.Variables,
	module string,
	callback func(uuid.UUID, map[string]cacao.Variables)) error {
	return nil
}

func (executer *Executer) Pause(command CommandData, module string) error {

	return nil
}

func (executer *Executer) Resume(command CommandData, module string) error {

	return nil
}

func (executer *Executer) Kill(command CommandData, module string) error {

	return nil
}
