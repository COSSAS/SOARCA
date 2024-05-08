package condition

import (
	"errors"
	"reflect"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
	"soarca/utils/stix/expression/comparison"
)

var component = reflect.TypeOf(Executor{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(comparison comparison.IComparison) *Executor {
	return &Executor{comparison: comparison}
}

type IExecuter interface {
	Execute(metadata execution.Metadata,
		step cacao.Step, variables cacao.Variables) (string, bool, error)
}

type Executor struct {
	comparison comparison.IComparison
}

func (executor *Executor) Execute(meta execution.Metadata, step cacao.Step, variables cacao.Variables) (string, bool, error) {

	if step.Type != cacao.StepTypeIfCondition {
		err := errors.New("the provided step type is not compatible with this executor")
		log.Error(err)
		return step.OnFailure, false, err
	}

	result, err := executor.comparison.Evaluate(step.Condition, variables)
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
