// Package workflow walks a playbook's workflow graph and runs its steps.
package workflow

import (
	"errors"
	"fmt"
	"reflect"

	"soarca/internal/logger"
	"soarca/internal/workflow/steps"
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
	"soarca/internal/reporting/cases"
	"soarca/internal/reporting/reporter"
	"soarca/pkg/utils/guid"
	timeUtil "soarca/pkg/utils/time"

	t "time"

	"github.com/google/uuid"
)

type Empty struct{}

var (
	component = reflect.TypeOf(Empty{}).PkgPath()
	log       *logger.Log
)

// Result is the outcome of walking one playbook.
type Result struct {
	RunId      uuid.UUID
	PlaybookId string
	Variables  cacao.Variables
}

// Walker walks one playbook's workflow graph. One per run: it holds per-run state.
type Walker interface {
	ExecuteAsync(playbook cacao.Playbook, results chan Result)
	Execute(playbook cacao.Playbook) (*Result, error)
}

// NewWalker builds a Walker for a single run.
type NewWalker func() Walker

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// New builds a walker. caseManager is optional; at most one may be supplied.
func New(action executors.IActionExecutor,
	subPlaybook executors.IPlaybookExecuter,
	condition executors.IConditionExecuter,
	guid guid.IGuid,
	reporter reporter.IWorkflowReporter,
	time timeUtil.ITime,
	caseManager ...cases.ICasesManager) *Walk {

	w := &Walk{
		action:      action,
		subPlaybook: subPlaybook,
		condition:   condition,
		guid:        guid,
		reporter:    reporter,
		time:        time,
	}
	if len(caseManager) > 0 {
		w.caseManager = caseManager[0]
	}
	return w
}

// Walk is a single playbook run in progress.
type Walk struct {
	playbook    cacao.Playbook
	result      Result
	action      executors.IActionExecutor
	subPlaybook executors.IPlaybookExecuter
	condition   executors.IConditionExecuter
	guid        guid.IGuid
	reporter    reporter.IWorkflowReporter
	caseManager cases.ICasesManager
	time        timeUtil.ITime
}

// ExecuteAsync starts a run and publishes its identity before walking the graph.
func (w *Walk) ExecuteAsync(playbook cacao.Playbook, results chan Result) {
	runId := w.guid.New()
	log.Debugf("Starting run %s for Playbook %s", runId, playbook.ID)

	result := Result{runId, playbook.ID, playbook.PlaybookVariables}
	w.result = result

	if results != nil {
		results <- result
	}

	_ = w.walk(playbook)
}

func (w *Walk) Execute(playbook cacao.Playbook) (*Result, error) {
	runId := w.guid.New()
	log.Debugf("Starting run %s for Playbook %s", runId, playbook.ID)
	w.result = Result{runId, playbook.ID, playbook.PlaybookVariables}

	err := w.walk(playbook)

	return &w.result, err
}

func (w *Walk) walk(playbook cacao.Playbook) error {

	w.playbook = playbook

	stepId := playbook.WorkflowStart

	// Start case correlation and get case ID to be used in playbook
	if w.caseManager != nil {
		startMetadata := w.newStepMetadata(stepId)

		caseIdVar := w.caseManager.AddToExistingOrCreateNew(startMetadata, playbook)
		playbook.PlaybookVariables.InsertOrReplace(caseIdVar)
		log.Info("case id is set to: ", caseIdVar.Value)
	}

	variables := cacao.NewVariables()
	variables.Merge(playbook.PlaybookVariables)

	// Reporting workflow instantiation
	w.reporter.ReportWorkflowStart(w.result.RunId, playbook, w.time.Now())

	outputVariables, err := w.ExecuteBranch(stepId, variables)

	w.result.Variables = outputVariables
	// Reporting workflow end
	w.reporter.ReportWorkflowEnd(w.result.RunId, playbook, err, w.time.Now())

	return err
}

// ExecuteBranch walks a branch of the workflow.
//
// Runs until it finds an End step or returns an error in case there are no valid next step.
func (w *Walk) ExecuteBranch(stepId string, scopeVariables cacao.Variables) (cacao.Variables, error) {
	playbook := w.playbook
	log.Debug("Running branch starting from ", stepId)

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

		// Note: likely (but not certainly) on_success and on_faliure will be reworked
		// to become workflow branching properties, with the addition of a success_condition
		// boolean evaluation at step level.
		// Effectively, we should thus only check for existance of on_completion, and
		// report run errors as such, not as playbook step failures - which will be handled
		// with upcoming said on_success, on_failure, and success_condition properties
		onCompletionStepId := currentStep.OnCompletion
		if onCompletionStepId == "" {
			onCompletionStepId = currentStep.OnSuccess
		}
		if onCompletionStepId == "" {
			onCompletionStepId = currentStep.OnFailure
		}
		if _, ok := playbook.Workflow[onCompletionStepId]; !ok {
			return cacao.NewVariables(), errors.New("empty completion step")
		}

		outputVariables, err := w.ExecuteStep(currentStep, scopeVariables)

		if err == nil {
			stepId = onCompletionStepId
			returnVariables.Merge(outputVariables)
			scopeVariables.Merge(outputVariables)
		} else {
			return cacao.NewVariables(), fmt.Errorf("playbook run failed at step [ %s ]. See step log for error information", stepId)
		}
	}

	return returnVariables, nil
}

// newStepMetadata builds the run.Metadata for one invocation of stepId,
// minting a fresh StepRunId each time it is called. Call it once per
// actual dispatch of a step (including once per while-loop iteration), never
// reuse a previously-built value across separate invocations.
func (w *Walk) newStepMetadata(stepId string) run.Metadata {
	return run.Metadata{
		RunId:      w.result.RunId,
		PlaybookId: w.result.PlaybookId,
		StepId:     stepId,
		StepRunId:  w.guid.New(),
	}
}

// ExecuteStep runs a single step within the workflow.
func (w *Walk) ExecuteStep(step cacao.Step, scopeVariables cacao.Variables) (cacao.Variables, error) {
	log.Debug("Running step type ", step.Type)

	log.Trace("Delay is set to: ", step.Delay)
	w.time.Sleep(t.Duration(step.Delay) * t.Millisecond)

	// Combine parent scope and Step variables
	variables := cacao.NewVariables()
	variables.Merge(scopeVariables)
	variables.Merge(step.StepVariables)

	switch step.Type {
	case cacao.StepTypeAction:
		metadata := w.newStepMetadata(step.ID)
		actionMetadata := executors.PlaybookStepMetadata{
			Step:      step,
			Targets:   w.playbook.TargetDefinitions,
			Auth:      w.playbook.AuthenticationInfoDefinitions,
			Agent:     w.playbook.AgentDefinitions[step.Agent],
			Variables: variables,
		}
		return w.action.Execute(metadata, actionMetadata)
	case cacao.StepTypePlaybookAction:
		metadata := w.newStepMetadata(step.ID)
		return w.subPlaybook.Execute(metadata, step, variables)
	case cacao.StepTypeIfCondition:
		return w.executeIfCondition(step, variables)
	case cacao.StepTypeWhileCondition:
		return w.executeLoop(step, variables)
	default:
		// NOTE: This currently silently handles unknown step types. Should we return an error instead?
		return cacao.NewVariables(), nil //errors.ErrUnsupported
	}
}

func (w *Walk) executeIfCondition(step cacao.Step,
	variables cacao.Variables) (cacao.Variables, error) {
	metadata := w.newStepMetadata(step.ID)
	stepId, branch, err := w.condition.Execute(metadata,
		executors.Context{Step: step, Variables: variables})
	if err != nil {
		return cacao.NewVariables(), err
	}
	if branch {
		return w.ExecuteBranch(stepId, variables)
	}
	return variables, nil
}

func (w *Walk) executeLoop(step cacao.Step,
	variables cacao.Variables) (cacao.Variables, error) {

	loop := true

	for loop {
		// A fresh StepRunId per iteration: each pass through the loop
		// re-evaluates this same while-condition step.
		metadata := w.newStepMetadata(step.ID)
		stepId, branch, err := w.condition.Execute(metadata,
			executors.Context{Step: step, Variables: variables})
		if err != nil {
			return cacao.NewVariables(), err
		}
		loop = branch

		if loop {
			branchVariables, err := w.ExecuteBranch(stepId, variables)
			if err != nil {
				return variables, err
			}
			variables.Merge(branchVariables)
		}

	}
	return variables, nil
}
