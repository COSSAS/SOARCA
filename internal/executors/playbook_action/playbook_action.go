package playbook_action

import (
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/controller/database"
	"soarca/internal/controller/decomposer_controller"
	"soarca/internal/reporter"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
	timeUtil "soarca/utils/time"
)

type PlaybookAction struct {
	decomposerController decomposer_controller.IController
	databaseController   database.IController
	reporter             reporter.IStepReporter
	time                 timeUtil.ITime
}

var component = reflect.TypeOf(PlaybookAction{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(controller decomposer_controller.IController,
	database database.IController, reporter reporter.IStepReporter, time timeUtil.ITime) *PlaybookAction {
	return &PlaybookAction{decomposerController: controller, databaseController: database, reporter: reporter, time: time}
}

func (playbookAction *PlaybookAction) Execute(metadata execution.Metadata,
	step cacao.Step,
	variables cacao.Variables) (cacao.Variables, error) {
	log.Trace(metadata.ExecutionId)

	playbookAction.reporter.ReportStepStart(metadata.ExecutionId, step, variables, playbookAction.time.Now())

	if step.Type != cacao.StepTypePlaybookAction {
		err := errors.New(fmt.Sprint("step type is not of type ", cacao.StepTypePlaybookAction))
		log.Error(err)
		playbookAction.reporter.ReportStepEnd(metadata.ExecutionId, step, cacao.NewVariables(), nil, playbookAction.time.Now())
		return cacao.NewVariables(), err
	}

	playbookRepo := playbookAction.databaseController.GetDatabaseInstance()
	decomposer := playbookAction.decomposerController.NewDecomposer()

	playbook, err := playbookRepo.Read(step.PlaybookID)
	if err != nil {
		log.Error("failed loading the playbook from the repository in playbook action")
		return cacao.NewVariables(), err
	}

	playbook.PlaybookVariables.Merge(variables)

	details, err := decomposer.Execute(playbook)
	if err != nil {
		err = errors.New(fmt.Sprint("execution of playbook failed with error: ", err))
		log.Error(err)
		playbookAction.reporter.ReportStepEnd(metadata.ExecutionId, step, details.Variables, err, playbookAction.time.Now())
		return cacao.NewVariables(), err
	}
	playbookAction.reporter.ReportStepEnd(metadata.ExecutionId, step, details.Variables, nil, playbookAction.time.Now())
	return details.Variables, nil

}
