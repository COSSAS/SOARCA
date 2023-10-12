package executer

type CommandData struct {
	id       string
	command  string
	isBase64 bool
}

type Variable struct{}
type AuthInfo struct{}
type Target struct{}
type Module struct{}

type IExecuter interface {
	Execute(command CommandData, variables []Variable, module Module, callback func(outputVariables []Variable)) error
	Pause(command CommandData, module Module) error
	Resume(command CommandData, module Module) error
	Kill(command CommandData, module Module) error
}

type Executer struct {
}

func (executer Executer) Execute(command CommandData, variables []Variable, module Module, callback func(outputVariables []Variable)) error {
	command = CommandData{}
	command.command = ""
	command.id = ""
	command.isBase64 = true

	return nil
}

func (executer Executer) Pause(command CommandData, module Module) error {

	return nil
}

func (executer Executer) Resume(command CommandData, module Module) error {

	return nil
}

func (executer Executer) Kill(command CommandData, module Module) error {

	return nil
}
