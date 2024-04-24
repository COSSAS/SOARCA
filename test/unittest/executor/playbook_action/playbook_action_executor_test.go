package playbook_action_executor_test

import (
	"testing"

	"soarca/internal/decomposer"
	"soarca/internal/executors/playbook_action"
	mock_database_controller "soarca/test/unittest/mocks/mock_controller/database"
	mock_decomposer_controller "soarca/test/unittest/mocks/mock_controller/decomposer"
	"soarca/test/unittest/mocks/mock_decomposer"
	"soarca/test/unittest/mocks/mock_reporter"
	mocks_playbook_test "soarca/test/unittest/mocks/playbook"

	"soarca/models/cacao"
	"soarca/models/execution"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestExecutePlaybook(t *testing.T) {

	playbookRepoMock := new(mocks_playbook_test.MockPlaybook)
	mockDecomposer := new(mock_decomposer.Mock_Decomposer)
	mock_reporter := new(mock_reporter.Mock_Reporter)

	controller := new(mock_decomposer_controller.Mock_Controller)
	database := new(mock_database_controller.Mock_Controller)

	executerObject := playbook_action.New(controller, database, mock_reporter)
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

	database.On("GetDatabaseInstance").Return(playbookRepoMock)
	controller.On("NewDecomposer").Return(mockDecomposer)
	mock_reporter.On("ReportStep", executionId, step, cacao.NewVariables(returnedVariables), nil).Return()

	playbook := cacao.Playbook{ID: playbookId, PlaybookVariables: cacao.NewVariables(initialVariables)}
	playbookRepoMock.On("Read", playbookId).Return(playbook, nil)
	details := decomposer.ExecutionDetails{ExecutionId: executionId,
		PlaybookId: playbookId,
		Variables:  cacao.NewVariables(returnedVariables)}

	playbook2 := cacao.Playbook{ID: playbookId, PlaybookVariables: cacao.NewVariables(expectedVariables)}
	mockDecomposer.On("Execute", playbook2).Return(&details, nil)

	results, err := executerObject.Execute(metadata, step, cacao.NewVariables(addedVariables))

	mock_reporter.AssertExpectations(t)
	assert.Equal(t, err, nil)
	assert.Equal(t, results, cacao.NewVariables(returnedVariables))

}
