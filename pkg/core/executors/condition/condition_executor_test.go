package condition

import (
	"errors"
	"soarca/models/cacao"
	"soarca/models/execution"
	"soarca/test/unittest/mocks/mock_reporter"
	mock_stix "soarca/test/unittest/mocks/mock_utils/stix"
	mock_time "soarca/test/unittest/mocks/mock_utils/time"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestExecuteConditionTrue(t *testing.T) {
	mock_stix := new(mock_stix.MockStix)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

	conditionExecutior := New(mock_stix, mock_reporter, mock_time)

	executionId := uuid.New()

	meta := execution.Metadata{ExecutionId: executionId,
		PlaybookId: "1",
		StepId:     "2"}

	step := cacao.Step{Type: cacao.StepTypeIfCondition,
		Condition: "a = a",
		OnTrue:    "3",
		OnFalse:   "4"}
	vars := cacao.NewVariables()

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	mock_reporter.On("ReportStepStart", executionId, step, vars, timeNow)
	mock_stix.On("Evaluate", "a = a", vars).Return(true, nil)
	mock_reporter.On("ReportStepEnd", executionId, step, vars, nil, timeNow)

	nextStepId, goToBranch, err := conditionExecutior.Execute(meta, step, vars)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, goToBranch)
	assert.Equal(t, "3", nextStepId)

	mock_reporter.AssertExpectations(t)
	mock_stix.AssertExpectations(t)
	mock_time.AssertExpectations(t)
}

func TestExecuteConditionFalse(t *testing.T) {
	mock_stix := new(mock_stix.MockStix)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

	conditionExecutior := New(mock_stix, mock_reporter, mock_time)

	executionId := uuid.New()

	meta := execution.Metadata{ExecutionId: executionId,
		PlaybookId: "1",
		StepId:     "2"}

	step := cacao.Step{Type: cacao.StepTypeIfCondition,
		Condition: "a = a",
		OnTrue:    "3",
		OnFalse:   "4"}
	vars := cacao.NewVariables()

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	mock_reporter.On("ReportStepStart", executionId, step, vars, timeNow)
	mock_stix.On("Evaluate", "a = a", vars).Return(false, nil)
	mock_reporter.On("ReportStepEnd", executionId, step, vars, nil, timeNow)

	nextStepId, goToBranch, err := conditionExecutior.Execute(meta, step, vars)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, goToBranch)
	assert.Equal(t, "4", nextStepId)

	mock_reporter.AssertExpectations(t)
	mock_stix.AssertExpectations(t)
	mock_time.AssertExpectations(t)
}

func TestExecuteConditionError(t *testing.T) {
	mock_stix := new(mock_stix.MockStix)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

	conditionExecutior := New(mock_stix, mock_reporter, mock_time)

	executionId := uuid.New()

	meta := execution.Metadata{ExecutionId: executionId,
		PlaybookId: "1",
		StepId:     "2"}

	step := cacao.Step{Type: cacao.StepTypeIfCondition,
		Condition: "a = a",
		OnTrue:    "3",
		OnFalse:   "4"}
	vars := cacao.NewVariables()

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	evaluationError := errors.New("some ds error")

	mock_reporter.On("ReportStepStart", executionId, step, vars, timeNow).Return()
	mock_stix.On("Evaluate", "a = a", vars).Return(false, evaluationError)
	mock_reporter.On("ReportStepEnd", executionId, step, vars, evaluationError, timeNow).Return()

	nextStepId, goToBranch, err := conditionExecutior.Execute(meta, step, vars)
	assert.Equal(t, evaluationError, err)
	assert.Equal(t, false, goToBranch)
	assert.Equal(t, "", nextStepId)

	mock_reporter.AssertExpectations(t)
	mock_stix.AssertExpectations(t)
	mock_time.AssertExpectations(t)
}
