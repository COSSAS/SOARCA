package playbook_action

import (
	"testing"
	"time"

	"soarca/pkg/core/decomposer"
	mock_database_controller "soarca/test/unittest/mocks/mock_controller/database"
	mock_decomposer_controller "soarca/test/unittest/mocks/mock_controller/decomposer"
	"soarca/test/unittest/mocks/mock_decomposer"
	mocks_playbook_test "soarca/test/unittest/mocks/mock_playbook_database"
	"soarca/test/unittest/mocks/mock_reporter"
	mock_time "soarca/test/unittest/mocks/mock_utils/time"

	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestExecutePlaybook(t *testing.T) {

	playbookRepoMock := new(mocks_playbook_test.MockPlaybook)
	mockDecomposer := new(mock_decomposer.Mock_Decomposer)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

	controller := new(mock_decomposer_controller.Mock_Controller)
	database := new(mock_database_controller.Mock_Controller)

	executorObject := New(controller, database, mock_reporter, mock_time)
	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId := "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	stepId := "step--81eff59f-d084-4324-9e0a-59e353dbd28f"

	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId, StepId: stepId}

	initialVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	addedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing2",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing2",
	}

	returnedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing2",
	}

	step := cacao.Step{
		Type:        cacao.StepTypePlaybookAction,
		Name:        "Playbook action test",
		ID:          stepId,
		Description: "",
		PlaybookID:  playbookId,
	}

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	database.On("GetDatabaseInstance").Return(playbookRepoMock)
	controller.On("NewDecomposer").Return(mockDecomposer)

	mock_reporter.On("ReportStepStart", executionId, step, cacao.NewVariables(addedVariables), timeNow).Return()
	mock_reporter.On("ReportStepEnd", executionId, step, cacao.NewVariables(returnedVariables), nil, timeNow).Return()

	playbook := cacao.Playbook{ID: playbookId, PlaybookVariables: cacao.NewVariables(initialVariables)}
	playbookRepoMock.On("Read", playbookId).Return(playbook, nil)
	details := decomposer.ExecutionDetails{ExecutionId: executionId,
		PlaybookId: playbookId,
		Variables:  cacao.NewVariables(returnedVariables)}

	playbook2 := cacao.Playbook{ID: playbookId, PlaybookVariables: cacao.NewVariables(expectedVariables)}

	mockDecomposer.On("Execute", playbook2).Return(&details, nil)

	results, err := executorObject.Execute(metadata, step, cacao.NewVariables(addedVariables))

	mockDecomposer.AssertExpectations(t)
	mock_reporter.AssertExpectations(t)
	mock_time.AssertExpectations(t)
	assert.Equal(t, err, nil)
	assert.Equal(t, results, cacao.NewVariables(returnedVariables))

}
