package decomposer

import (
	"errors"
	"fmt"
	"reflect"

	"soarca/internal/executors"
	"soarca/internal/executors/action"
	"soarca/internal/executors/condition"
	"soarca/internal/guid"
	"soarca/internal/reporter"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
	"soarca/utils/time"

	t "time"

	"github.com/google/uuid"
)

type Empty struct{}

var (
	component = reflect.TypeOf(Empty{}).PkgPath()
	log       *logger.Log
)

type ExecutionDetails struct {
	ExecutionId uuid.UUID
	PlaybookId  string
	Variables   cacao.Variables
}

type IDecomposer interface {
	ExecuteAsync(playbook cacao.Playbook, detailsch chan ExecutionDetails)
	Execute(playbook cacao.Playbook) (*ExecutionDetails, error)
}

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(actionExecutor action.IExecuter,
	playbookActionExecutor executors.IPlaybookExecuter,
	condition condition.IExecuter,
	guid guid.IGuid,
	reporter reporter.IWorkflowReporter,
	time time.ITime) *Decomposer {

	return &Decomposer{actionExecutor: actionExecutor,
		playbookActionExecutor: playbookActionExecutor,
		conditionExecutor:      condition,
		guid:                   guid,
		reporter:               reporter,
		time:                   time,
	}
}

type Decomposer struct {
	playbook               cacao.Playbook
	details                ExecutionDetails
	actionExecutor         action.IExecuter
	playbookActionExecutor executors.IPlaybookExecuter
	conditionExecutor      condition.IExecuter
	guid                   guid.IGuid
	reporter               reporter.IWorkflowReporter
	time                   time.ITime
}

// Execute a Playbook
func (decomposer *Decomposer) ExecuteAsync(playbook cacao.Playbook, detailsch chan ExecutionDetails) {
	executionId := decomposer.guid.New()
	log.Debugf("Starting execution %s for Playbook %s", executionId, playbook.ID)

	details := ExecutionDetails{executionId, playbook.ID, playbook.PlaybookVariables}
	decomposer.details = details

	if detailsch != nil {
		detailsch <- details
	}

	decomposer.execute(playbook)

}

func (decomposer *Decomposer) Execute(playbook cacao.Playbook) (*ExecutionDetails, error) {

	executionId := decomposer.guid.New()
	log.Debugf("Starting execution %s for Playbook %s", executionId, playbook.ID)
	decomposer.details = ExecutionDetails{executionId, playbook.ID, playbook.PlaybookVariables}

	err := decomposer.execute(playbook)

	return &decomposer.details, err

}

func (decomposer *Decomposer) execute(playbook cacao.Playbook) error {

	decomposer.playbook = playbook

	stepId := playbook.WorkflowStart
	variables := cacao.NewVariables()
	variables.Merge(playbook.PlaybookVariables)

	// Reporting workflow instantiation
	decomposer.reporter.ReportWorkflowStart(decomposer.details.ExecutionId, playbook)

	outputVariables, err := decomposer.ExecuteBranch(stepId, variables)

	decomposer.details.Variables = outputVariables
	// Reporting workflow end
	decomposer.reporter.ReportWorkflowEnd(decomposer.details.ExecutionId, playbook, err)

	return err
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

	log.Trace("Delay is set to: ", step.Delay)
	decomposer.time.Sleep(t.Duration(step.Delay) * t.Second)

	// Combine parent scope and Step variables
	variables := cacao.NewVariables()
	variables.Merge(scopeVariables)
	variables.Merge(step.StepVariables)

	metadata := execution.Metadata{
		ExecutionId: decomposer.details.ExecutionId,
		PlaybookId:  decomposer.details.PlaybookId,
		StepId:      step.ID,
	}

	switch step.Type {
	case cacao.StepTypeAction:
		actionMetadata := action.PlaybookStepMetadata{
			Step:      step,
			Targets:   decomposer.playbook.TargetDefinitions,
			Auth:      decomposer.playbook.AuthenticationInfoDefinitions,
			Agent:     decomposer.playbook.AgentDefinitions[step.Agent],
			Variables: variables,
		}
		return decomposer.actionExecutor.Execute(metadata, actionMetadata)
	case cacao.StepTypePlaybookAction:
		return decomposer.playbookActionExecutor.Execute(metadata, step, variables)
	case cacao.StepTypeIfCondition:
		stepId, branch, err := decomposer.conditionExecutor.Execute(metadata, step, variables)
		if err != nil {
			return cacao.NewVariables(), err
		}
		if branch {
			return decomposer.ExecuteBranch(stepId, variables)
		}
		return variables, nil
	default:
		// NOTE: This currently silently handles unknown step types. Should we return an error instead?
		return cacao.NewVariables(), nil
	}
}
