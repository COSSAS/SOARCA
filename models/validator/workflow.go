package validator

import (
	"errors"
	"fmt"
	"soarca/logger"
	"soarca/models/cacao"
)

func init() {
	log = logger.Logger(component, logger.Trace, "", logger.Json)
}

func IsSafeCacaoWorkflow(playbook *cacao.Playbook) error {

	// Workflow exception handled?
	workflowException := playbook.WorkflowException
	if workflowException == "" {
		log.Warn("workflow exception not implemented")
	}

	workflowStart := playbook.WorkflowStart

	_, ok := playbook.Workflow[workflowStart]
	if !ok {
		return errors.New("start step " + workflowStart + " not found in workflow")
	}

	// Check if all steps and data exist.
	// NOTE: it DOES NOT check for variables
	for stepId := range playbook.Workflow {
		err := isSafeWorkflowStep(playbook, stepId)
		if err != nil {
			//log.Error("step " + stepId + " validation raised an error")
			log.Error(err)

			return err
		}
	}

	// Check if there are no infinite loops in the workflow
	if err := checkAllWorkflowBranchesEnd(playbook); err != nil {
		return err
	}

	return nil
}

// Given one step id, checks if its execution is safe
// with respect to all properties being present in the playbook.
// All checks are in O(1) as it's all dictionary key lookups
func isSafeWorkflowStep(playbook *cacao.Playbook, stepId string) error {

	workflow := playbook.Workflow
	step := workflow[stepId]

	// checkSubStepsExist
	if err := checkStepSubStepsExist(workflow, step); err != nil {
		return err
	}
	// Check agents exist
	if err := checkStepAgentsExist(playbook, step); err != nil {
		return err
	}
	// Check targets exist
	if err := checkStepTargetsExist(playbook, step); err != nil {
		return err
	}
	// Check auth info exists
	if err := checkStepAuthInfoExist(playbook, step); err != nil {
		return err
	}

	return nil
}

func checkStepAgentsExist(playbook *cacao.Playbook, s cacao.Step) error {
	if s.Agent != "" {
		if _, ok := playbook.AgentDefinitions[s.Agent]; !ok {
			return errors.New("agent " + s.Agent +
				"not found in agent_definitions")
		}
	}
	return nil
}

func checkStepTargetsExist(playbook *cacao.Playbook, s cacao.Step) error {
	if len(s.Targets) > 0 {
		for _, t := range s.Targets {
			if _, ok := playbook.TargetDefinitions[t]; !ok {
				return errors.New("target " + t +
					"not found in target_definitions")
			}
		}
	}
	return nil
}

func checkStepAuthInfoExist(playbook *cacao.Playbook, s cacao.Step) error {
	if s.AuthenticationInfo != "" {
		if _, ok :=
			playbook.AuthenticationInfoDefinitions[s.AuthenticationInfo]; !ok {
			return errors.New("authenticaiton_info " +
				s.AuthenticationInfo +
				"not found in authentication_info_definitions")
		}
	}
	return nil
}

// TODO change to explicit vars
func checkStepSubStepsExist(workflow cacao.Workflow, step cacao.Step) error {

	if _, ok := workflow[step.OnCompletion]; step.OnCompletion != "" && !ok {
		return errors.New("step " + step.OnCompletion + " does not exist")
	}
	if _, ok := workflow[step.OnSuccess]; step.OnSuccess != "" && !ok {
		return errors.New("step " + step.OnSuccess + " does not exist")
	}
	if _, ok := workflow[step.OnFailure]; step.OnFailure != "" && !ok {
		return errors.New("step " + step.OnFailure + " does not exist")
	}
	if _, ok := workflow[step.OnTrue]; step.OnTrue != "" && !ok {
		return errors.New("step " + step.OnTrue + " does not exist")
	}
	if _, ok := workflow[step.OnFalse]; step.OnFalse != "" && !ok {
		return errors.New("step " + step.OnFalse + " does not exist")
	}

	if len(step.NextSteps) > 0 {
		for _, id := range step.NextSteps {
			if _, ok := workflow[id]; id != "" && !ok {
				return errors.New("step " + id + " does not exist")
			}
		}
	}
	if len(step.Cases) > 0 {
		for _, i := range step.Cases {
			if _, ok := workflow[i]; i != "" && !ok {
				return errors.New("step " + i + " does not exist")
			}
		}
	}
	return nil
}

// Wrapper for recursive func allBranchesEnd
func checkAllWorkflowBranchesEnd(playbook *cacao.Playbook) error {
	workflow := playbook.Workflow
	ss := playbook.WorkflowStart

	err := allBranchesEnd(workflow, ss, make(map[string]struct{}, 1))

	if err != nil {
		return err
	}

	return nil
}

// Navigates a CACAO workflow object recursively as depth-first tree
// on possible branches of the workflow execution.
//
//	branchSequence parameter "branching sequence" collects all the steps ID
//	 	in the current branch and copies its values to
//		the next recursive step in order to check if loops are present.
func allBranchesEnd(workflow cacao.Workflow, id string, branchSequence map[string]struct{}) error {

	// current branch sequence is copy of passed branch sequence
	currentBranchSequence := make(map[string]struct{}, len(branchSequence))
	for i, v := range branchSequence {
		currentBranchSequence[i] = v
	}
	step := workflow[id]
	currentBranchSequence[id] = struct{}{}

	children := map[string]cacao.Step{}

	// Fetch all possible children steps (branches)
	if ocs := step.OnCompletion; ocs != "" {
		children[ocs] = workflow[ocs]
	}
	if oss := step.OnSuccess; oss != "" {
		children[oss] = workflow[oss]
	}
	if ofs := step.OnFailure; ofs != "" {
		children[ofs] = workflow[ofs]
	}
	if ot := step.OnTrue; ot != "" {
		children[ot] = workflow[ot]
	}
	if of := step.OnFalse; of != "" {
		children[of] = workflow[of]
	}
	if nss := step.NextSteps; len(nss) > 0 {
		for _, stepId := range nss {
			children[stepId] = workflow[stepId]
		}
	}
	if len(step.Cases) > 0 {
		for _, stepId := range step.Cases {
			children[stepId] = workflow[stepId]
		}
	}

	// Recursion end
	if len(children) == 0 {
		if step.Type == cacao.StepTypeEnd {
			// Leaf
			return nil
		} else {
			// Should never reach this condition due to schema validation
			return errors.New("step with no branches is not an end step")
		}
	}

	for child := range children {
		// If the children of the current node contain one node
		// that was already explored, as seen in currentBranchStep (current branch steps)
		// then there is an infinite loop
		if _, e := currentBranchSequence[child]; e {
			currentBranchSequence["infinite#"+child] = struct{}{}
			return errors.New("worflow seems to loop on branch sequence " + fmt.Sprint(currentBranchSequence))
		}
		// Call recursive workflow exploration on children
		err := allBranchesEnd(workflow, child, currentBranchSequence)
		if err != nil {
			return err
		}
	}
	return nil
}
