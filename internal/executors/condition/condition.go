package condition

import (
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/reporter"
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

func New(comparison comparison.IComparison,
	reporter reporter.IStepReporter) *Executor {
	return &Executor{comparison: comparison,
		reporter: reporter}
}

type IExecuter interface {
	Execute(metadata execution.Metadata,
		step cacao.Step, variables cacao.Variables) (string, bool, error)
}

type Executor struct {
	comparison comparison.IComparison
	reporter   reporter.IStepReporter
}

func (executor *Executor) Execute(meta execution.Metadata, step cacao.Step, variables cacao.Variables) (string, bool, error) {

	if step.Type != cacao.StepTypeIfCondition {
		err := errors.New("the provided step type is not compatible with this executor")
		log.Error(err)
		return step.OnFailure, false, err
	}

	executor.reporter.ReportStepStart(meta.ExecutionId, step, variables)

	result, err := executor.comparison.Evaluate(step.Condition, variables)

	// We are reporting early to not have double reporting
	executor.reporter.ReportStepEnd(meta.ExecutionId,
		step,
		variables,
		err)

	if err != nil {
		log.Error(err)
		return "", false, err
	}

	log.Debug("the result was: ", fmt.Sprint(result))

	if result {
		if step.OnTrue != "" {
			log.Trace("will return on true step ", step.OnTrue)
			return step.OnTrue, true, nil
		}
	} else {
		if step.OnFalse != "" {
			log.Trace("will return on false step ", step.OnFalse)
			return step.OnFalse, true, nil
		}
	}
	log.Trace("will return on completion step ", step.OnCompletion)

	return step.OnCompletion, false, nil
}
