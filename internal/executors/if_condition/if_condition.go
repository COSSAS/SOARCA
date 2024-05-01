package ifcondition

import (
	"errors"
	"reflect"
	"soarca/internal/capability"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
)

var component = reflect.TypeOf(Executor{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(capabilities map[string]capability.ICapability) *Executor {
	var instance = Executor{}
	instance.capabilities = capabilities
	return &instance
}

type IExecuter interface {
	Execute(metadata execution.Metadata,
		step cacao.Step) (string, error)
}

type Executor struct {
	capabilities map[string]capability.ICapability
}

func (executor *Executor) Execute(meta execution.Metadata, step cacao.Step) (string, error) {

	if step.Type != cacao.StepTypeIfCondition {
		err := errors.New("the provided step type is not compatible with this executor")
		log.Error(err)
		return step.OnFailure, err
	}
	return "", nil
}
