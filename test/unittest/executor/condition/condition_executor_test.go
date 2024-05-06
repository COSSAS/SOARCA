package condition_test

import (
	"soarca/internal/executors/condition"
	"soarca/models/cacao"
	"soarca/models/execution"
	mock_stix "soarca/test/unittest/mocks/mock_utils/stix"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestExecuteConditionTrue(t *testing.T) {
	mock_stix := new(mock_stix.MockStix)

	conditionExecutior := condition.New(mock_stix)

	executionId := uuid.New()

	meta := execution.Metadata{ExecutionId: executionId,
		PlaybookId: "1",
		StepId:     "2"}

	step := cacao.Step{Type: cacao.StepTypeIfCondition,
		Condition: "a = a",
		OnTrue:    "3",
		OnFalse:   "4"}
	vars := cacao.NewVariables()

	mock_stix.On("Evaluate", "a = a", vars).Return(true, nil)

	nextStepId, goToBranch, err := conditionExecutior.Execute(meta, step, vars)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, goToBranch)
	assert.Equal(t, "3", nextStepId)

}

func TestExecuteConditionFalse(t *testing.T) {
	mock_stix := new(mock_stix.MockStix)

	conditionExecutior := condition.New(mock_stix)

	executionId := uuid.New()

	meta := execution.Metadata{ExecutionId: executionId,
		PlaybookId: "1",
		StepId:     "2"}

	step := cacao.Step{Type: cacao.StepTypeIfCondition,
		Condition: "a = a",
		OnTrue:    "3",
		OnFalse:   "4"}
	vars := cacao.NewVariables()

	mock_stix.On("Evaluate", "a = a", vars).Return(true, nil)

	nextStepId, goToBranch, err := conditionExecutior.Execute(meta, step, vars)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, goToBranch)
	assert.Equal(t, "3", nextStepId)

}
