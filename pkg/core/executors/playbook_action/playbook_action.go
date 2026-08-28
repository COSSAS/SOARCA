package playbook_action

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/controller/decomposer_controller"
	"soarca/internal/logger"
	"soarca/internal/storage"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/reporting/reporter"
	timeUtil "soarca/pkg/utils/time"
)

type PlaybookAction struct {
	decomposerController decomposer_controller.IController
	playbookStore        storage.PlaybookStore
	reporter             reporter.IStepReporter
	time                 timeUtil.ITime
}

var component = reflect.TypeOf(PlaybookAction{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(controller decomposer_controller.IController, playbookStore storage.PlaybookStore, reporter reporter.IStepReporter, time timeUtil.ITime) *PlaybookAction {
	return &PlaybookAction{decomposerController: controller, playbookStore: playbookStore, reporter: reporter, time: time}
}

func (playbookAction *PlaybookAction) Execute(metadata execution.Metadata,
	step cacao.Step,
	variables cacao.Variables) (cacao.Variables, error) {
	log.Trace(metadata.ExecutionId)

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

	decomposer := playbookAction.decomposerController.NewDecomposer()
	details, err := decomposer.Execute(playbook)
	if err != nil {
		err = errors.New(fmt.Sprint("execution of playbook failed with error: ", err))
		log.Error(err)
		reportVars = details.Variables
		return cacao.NewVariables(), err
	}
	reportVars = details.Variables
	return details.Variables, nil

}
