package condition

import (
	"errors"
	"reflect"
	"soarca/internal/capability"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
	"soarca/utils/stix"
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
		step cacao.Step, variables cacao.Variables) (string, bool, error)
}

type Executor struct {
	capabilities map[string]capability.ICapability
}

func (executor *Executor) Execute(meta execution.Metadata, step cacao.Step, variables cacao.Variables) (string, bool, error) {

	if step.Type != cacao.StepTypeIfCondition {
		err := errors.New("the provided step type is not compatible with this executor")
		log.Error(err)
		return step.OnFailure, false, err
	}

	result, err := stix.Evaluate(step.Condition, step.StepVariables)
	if err != nil {
		log.Error(err)
		return "", false, err
	}

	if result {
		if step.OnTrue != "" {
			log.Trace("")
			return step.OnTrue, true, nil
		}
	} else {
		if step.OnFalse != "" {
			return step.OnFalse, true, nil
		}
	}

	return step.OnCompletion, false, nil
}
