package decomposer

import (
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/executer"
	"soarca/internal/guid"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"

	"github.com/google/uuid"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

type ExecutionDetails struct {
	ExecutionId uuid.UUID
	PlaybookId  string
	Variables   cacao.Variables
}

type IDecomposer interface {
	Execute(playbook cacao.Playbook) (*ExecutionDetails, error)
}

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
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
	playbook cacao.Playbook
	details  ExecutionDetails
	executor executer.IExecuter
	guid     guid.IGuid
}

// Execute a Playbook
func (decomposer *Decomposer) Execute(playbook cacao.Playbook) (*ExecutionDetails, error) {
	executionId := decomposer.guid.New()
	log.Debugf("Starting execution %s for Playbook %s", executionId, playbook.ID)

	decomposer.details = ExecutionDetails{executionId, playbook.ID, playbook.PlaybookVariables}
	decomposer.playbook = playbook

	stepId := playbook.WorkflowStart
	variables := cacao.NewVariables()
	variables.Merge(playbook.PlaybookVariables)

	outputVariables, err := decomposer.ExecuteBranch(stepId, variables)

	decomposer.details.Variables = outputVariables
	return &decomposer.details, err
}

// Execute a Workflow branch of a Playbook
//
// Runs until it find an End step or returns an error in case there are no valid next step.
func (decomposer *Decomposer) ExecuteBranch(stepId string, scopeVariables cacao.Variables) (cacao.Variables, error) {
	playbook := decomposer.playbook
	log.Debug("Executing branch starting from ", stepId)

	returnVariables := cacao.NewVariables()

	for {
		currentStep, ok := playbook.Workflow[stepId]
		if !ok {
			return cacao.NewVariables(), fmt.Errorf("step with id %s not found", stepId)
		}

		log.Debug("Executing step ", stepId)

		if currentStep.Type == "end" {
			break
		}

		onSuccessStepId := currentStep.OnSuccess
		if onSuccessStepId == "" {
			onSuccessStepId = currentStep.OnCompletion
		}
		if _, ok := playbook.Workflow[onSuccessStepId]; !ok {
			return cacao.NewVariables(), errors.New("empty success step")
		}

		onFailureStepId := currentStep.OnFailure
		if onFailureStepId == "" {
			onFailureStepId = currentStep.OnCompletion
		}
		if _, ok := playbook.Workflow[onFailureStepId]; !ok {
			return cacao.NewVariables(), errors.New("empty failure step")
		}

		outputVariables, err := decomposer.ExecuteStep(currentStep, scopeVariables)

		if err == nil {
			stepId = onSuccessStepId
			returnVariables.Merge(outputVariables)
			scopeVariables.Merge(outputVariables)
		} else {
			stepId = onFailureStepId
		}
	}

	return returnVariables, nil
}

// Execute a single Step within a Workflow
func (decomposer *Decomposer) ExecuteStep(step cacao.Step, scopeVariables cacao.Variables) (cacao.Variables, error) {
	log.Debug("Executing step type ", step.Type)

	// Combine parent scope and Step variables
	variables := cacao.NewVariables()
	variables.Merge(scopeVariables)
	variables.Merge(step.StepVariables)

	switch step.Type {
	case "action":
		return decomposer.ExecuteActionStep(step, variables)
	default:
		// NOTE: This currently silently handles unknown step types. Should we return an error instead?
		return cacao.NewVariables(), nil
	}
}

// Execute a Step of type Action
func (decomposer *Decomposer) ExecuteActionStep(step cacao.Step, scopeVariables cacao.Variables) (cacao.Variables, error) {
	log.Debug("Executing action step")

	agent := decomposer.playbook.AgentDefinitions[step.Agent]
	returnVariables := cacao.NewVariables()

	for _, command := range step.Commands {
		// NOTE: This assumes we want to run Command for every Target individually.
		//       Is that something we want to enforce or leave up to the capability?
		for _, element := range step.Targets {
			target := decomposer.playbook.TargetDefinitions[element]
			// NOTE: What about Agent authentication?
			auth := decomposer.playbook.AuthenticationInfoDefinitions[target.AuthInfoIdentifier]

			meta := execution.Metadata{
				ExecutionId: decomposer.details.ExecutionId,
				PlaybookId:  decomposer.playbook.ID,
				StepId:      step.ID,
			}

			_, outputVariables, err := decomposer.executor.Execute(
				meta,
				command,
				auth,
				target,
				scopeVariables,
				agent)

			if len(step.OutArgs) > 0 {
				// If OutArgs is set, only update execution args that are explicitly referenced
				outputVariables = outputVariables.Select(step.OutArgs)
			}

			log.Tracef("Step output: %v", outputVariables)
			returnVariables.Merge(outputVariables)
			scopeVariables.Merge(outputVariables)

			if err != nil {
				log.Error("Error executing Command")
				return cacao.NewVariables(), err
			} else {
				log.Debug("Command executed")
			}
		}
	}

	return returnVariables, nil
}
