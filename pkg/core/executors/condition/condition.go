package condition

import (
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/core/executors"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/reporter"
	"soarca/pkg/utils/stix/expression/comparison"
	timeUtil "soarca/pkg/utils/time"
)

var component = reflect.TypeOf(Executor{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(comparison comparison.IComparison,
	reporter reporter.IStepReporter, time timeUtil.ITime) *Executor {
	return &Executor{comparison: comparison,
		reporter: reporter, time: time}
}

type IExecutor interface {
	Execute(metadata execution.Metadata,
		step cacao.Step, variables cacao.Variables) (string, bool, error)
}

type Executor struct {
	comparison comparison.IComparison
	reporter   reporter.IStepReporter
	time       timeUtil.ITime
}

func (executor *Executor) Execute(meta execution.Metadata, stepContext executors.Context) (string, bool, error) {

	if !(stepContext.Step.Type == cacao.StepTypeIfCondition || stepContext.Step.Type == cacao.StepTypeWhileCondition) {
		err := errors.New("the provided step type is not compatible with this executor")
		log.Error(err)
		return stepContext.Step.OnFailure, false, err
	}

	executor.reporter.ReportStepStart(meta.ExecutionId, stepContext.Step, stepContext.Variables, executor.time.Now())

	var err error
	defer func() {
		executor.reporter.ReportStepEnd(meta.ExecutionId, stepContext.Step, stepContext.Variables, err, executor.time.Now())
	}()
	nextStepId, branch, err := executor.evaluate(stepContext)
	return nextStepId, branch, err
}

func (executor *Executor) evaluate(stepContext executors.Context) (string, bool, error) {
	result, err := executor.comparison.Evaluate(stepContext.Step.Condition,
		stepContext.Variables)
	if err != nil {
		log.Error(err)
		return "", false, err
	}

	log.Debug("the result was: ", fmt.Sprint(result))

	if result {
		if stepContext.Step.OnTrue != "" {
			log.Trace("will return on true step ", stepContext.Step.OnTrue)
			return stepContext.Step.OnTrue, true, nil
		}
	} else {
		if stepContext.Step.OnFalse != "" {
			log.Trace("will return on false step ", stepContext.Step.OnFalse)
			return stepContext.Step.OnFalse, true, nil
		}
	}
	log.Trace("will return on completion step ", stepContext.Step.OnCompletion)

	return stepContext.Step.OnCompletion, false, nil
}
