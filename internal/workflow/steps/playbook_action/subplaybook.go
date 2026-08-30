package playbook_action

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/logger"
	"soarca/internal/store"
	"soarca/internal/workflow"
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
	"soarca/internal/reporting/reporter"
	timeUtil "soarca/pkg/utils/time"
)

type PlaybookAction struct {
	newWalker     workflow.NewWalker
	playbookStore storage.PlaybookStore
	reporter      reporter.IStepReporter
	time          timeUtil.ITime
}

var component = reflect.TypeOf(PlaybookAction{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(newWalker workflow.NewWalker, playbookStore storage.PlaybookStore, reporter reporter.IStepReporter, time timeUtil.ITime) *PlaybookAction {
	return &PlaybookAction{newWalker: newWalker, playbookStore: playbookStore, reporter: reporter, time: time}
}

func (playbookAction *PlaybookAction) Execute(metadata run.Metadata,
	step cacao.Step,
	variables cacao.Variables) (cacao.Variables, error) {
	log.Trace(metadata.RunId)

	playbookAction.reporter.ReportStepStart(metadata, step, variables, playbookAction.time.Now())

	var reportVars = cacao.NewVariables()
	var err error
	defer func() {
		playbookAction.reporter.ReportStepEnd(metadata, step, reportVars, err, playbookAction.time.Now())
	}()

	if step.Type != cacao.StepTypePlaybookAction {
		err := errors.New(fmt.Sprint("step type is not of type ", cacao.StepTypePlaybookAction))
		log.Error(err)
		return cacao.NewVariables(), err
	}

	playbook, err := playbookAction.playbookStore.Get(context.Background(), step.PlaybookID)
	if err != nil {
		log.Error("failed loading the playbook from storage in playbook action")
		return cacao.NewVariables(), err
	}

	playbook.PlaybookVariables.Merge(variables)

	walker := playbookAction.newWalker()
	result, err := walker.Execute(playbook)
	if err != nil {
		err = errors.New(fmt.Sprint("sub-playbook run failed with error: ", err))
		log.Error(err)
		reportVars = result.Variables
		return cacao.NewVariables(), err
	}
	reportVars = result.Variables
	return result.Variables, nil

}
