package conversion

import (
	"os"
	"soarca/pkg/conversion"
	"soarca/pkg/models/cacao"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func loadPlaybook(t *testing.T, filename string) *cacao.Playbook {
	input, err := os.ReadFile(filename)
	assert.Nil(t, err)
	playbook, err := conversion.NewBpmnConverter().Convert(input, filename)
	assert.Nil(t, err)
	return playbook
}

func nextSteps(t *testing.T, step cacao.Step, playbook *cacao.Playbook) []cacao.Step {
	var steps []cacao.Step
	for _, step_name := range step.NextSteps {
		steps = append(steps, findStep(t, step_name, playbook))
	}
	return steps
}
func findStep(t *testing.T, step_name string, playbook *cacao.Playbook) cacao.Step {
	step, ok := playbook.Workflow[step_name]
	assert.True(t, ok, "Could not find %s", step_name)
	return step
}
func findStepByName[S ~[]cacao.Step](t *testing.T, step_name string, steps S) *cacao.Step {
	for _, step := range steps {
		if step.Name == step_name {
			return &step
		}
	}
	assert.Fail(t, "Could not find name %s", step_name)
	return nil
}

func nextStep(t *testing.T, step cacao.Step, playbook *cacao.Playbook) cacao.Step {
	return findStep(t, step.OnCompletion, playbook)
}
func startStep(t *testing.T, playbook *cacao.Playbook) cacao.Step {
	return findStep(t, playbook.WorkflowStart, playbook)
}

func TestControlGatesConversion(t *testing.T) {
	playbook := loadPlaybook(t, "control_gates.bpmn")
	now := time.Now()
	assert.True(t, playbook.Created.Before(now))
	start := startStep(t, playbook)
	stepA := nextStep(t, start, playbook)
	assert.Equal(t, stepA.Name, "Task A")
	exclusive := nextStep(t, stepA, playbook)
	assert.True(t, exclusive.Type == cacao.StepTypeIfCondition)
	yesCase := findStep(t, exclusive.OnTrue, playbook)
	noCase := findStep(t, exclusive.OnFalse, playbook)
	assert.Equal(t, noCase.Name, "Task B")
	assert.Equal(t, yesCase.Name, "Task C")
	next := nextStep(t, noCase, playbook)
	assert.True(t, next.Type == cacao.StepTypeEnd)
	parallel := nextStep(t, yesCase, playbook)
	assert.True(t, parallel.Type == cacao.StepTypeParallel)
	nexts := nextSteps(t, parallel, playbook)
	stepD := findStepByName(t, "Task D", nexts)
	stepE := findStepByName(t, "Task E", nexts)
	next = nextStep(t, *stepD, playbook)
	assert.True(t, next.Type == cacao.StepTypeEnd)
	next = nextStep(t, *stepE, playbook)
	assert.True(t, next.Type == cacao.StepTypeEnd)
}

func TestSimpleSshConversion(t *testing.T) {
	playbook := loadPlaybook(t, "simple_ssh.bpmn")
	now := time.Now()
	assert.True(t, playbook.Created.Before(now))
	start := startStep(t, playbook)
	next := nextStep(t, start, playbook)
	assert.Equal(t, next.Name, "Execute command")
	next = nextStep(t, next, playbook)
	assert.Equal(t, next.Name, "Touch file")
	next = nextStep(t, next, playbook)
	assert.True(t, next.Type == cacao.StepTypeEnd)
}
