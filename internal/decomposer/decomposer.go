package decomposer

import (
	"errors"
	"reflect"
	"soarca/internal/executer"
	"soarca/internal/guid"
	"soarca/logger"
	"soarca/models/cacao"

	"github.com/google/uuid"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

type ExecutionDetails struct {
	ExecutionId uuid.UUID
	PlaybookId  string
	CurrentStep string
}

type IDecomposer interface {
	Execute(playbook cacao.Playbook) (*ExecutionDetails, error)
}

func init() {
	log = logger.Logger(component, logger.Trace, "", logger.Json)
}

func New(executor executer.IExecuter, guid guid.IGuid) *Decomposer {
	var instance = Decomposer{}
	if instance.executor == nil {
		instance.executor = executor
	}
	if instance.guid == nil {
		instance.guid = guid
	}
	return &instance
}

type Decomposer struct {
	playbooks cacao.Playbook
	details   ExecutionDetails
	executor  executer.IExecuter
	guid      guid.IGuid
}

func Callback(executionId uuid.UUID, outputVariables map[string]cacao.Variables) {

}

func (decomposer *Decomposer) Execute(playbook cacao.Playbook) (*ExecutionDetails, error) {
	var executionId = decomposer.guid.New()

	decomposer.details = ExecutionDetails{executionId, playbook.ID, ""}
	decomposer.playbooks = playbook

	var stepId = playbook.WorkflowStart

	for {
		if playbook.Workflow[stepId].OnCompletion == "" &&
			playbook.Workflow[stepId].Type == "end" ||
			playbook.Workflow[stepId].Type == "end" {
			break
		} else if playbook.Workflow[stepId].OnCompletion == "" {
			err := errors.New("empty on_completion_id")
			return &decomposer.details, err
		} else if _, ok := playbook.Workflow[playbook.Workflow[stepId].OnCompletion]; !ok {
			err := errors.New("on_completion_id key is not in workflows")
			return &decomposer.details, err
		}
		if len(playbook.Workflow[stepId].Commands) > 0 {
			for _, command := range playbook.Workflow[stepId].Commands {
				agent := playbook.AgentDefinitions[playbook.Workflow[stepId].Agent]

				for _, element := range playbook.Workflow[stepId].Targets {
					target := playbook.TargetDefinitions[element]
					auth := playbook.AuthenticationInfoDefinitions[target.AuthInfoIdentifier]

					var id, vars, _ = decomposer.executor.Execute(executionId,
						command,
						auth,
						target,
						playbook.Workflow[stepId].StepVariables,
						agent)
					log.Trace(id)
					log.Trace(vars)
				}
			}
		}
		stepId = playbook.Workflow[stepId].OnCompletion

	}

	return &decomposer.details, nil
}
