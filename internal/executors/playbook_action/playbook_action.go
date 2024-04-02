package playbook_action

import (
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/controller/database"
	"soarca/internal/controller/decomposer"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
)

const (
	resultVariable = "__soarca_playbook_action_result__"
	capabilityName = "playbook-action"
)

type PlaybookAction struct {
	decomposerController decomposer.IController
	databaseController   database.IController
}

var component = reflect.TypeOf(PlaybookAction{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(controller decomposer.IController,
	database database.IController) *PlaybookAction {
	return &PlaybookAction{decomposerController: controller, databaseController: database}
}

func (playbookAction *PlaybookAction) Execute(metadata execution.Metadata,
	step cacao.Step,
	variables cacao.Variables) (cacao.Variables, error) {
	log.Trace(metadata.ExecutionId)

	if step.Type != cacao.StepTypePlaybookAction {
		err := errors.New(fmt.Sprint("step type is not of type ", cacao.StepTypePlaybookAction))
		log.Error(err)
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
		return cacao.NewVariables(), err
	}
	return details.Variables, nil

}
