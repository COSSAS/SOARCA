package playbook_action

import (
	"testing"
	"time"

	"soarca/internal/workflow"
	mocks_playbook_test "soarca/test/unittest/mocks/mock_playbook_database"
	"soarca/test/unittest/mocks/mock_reporter"
	mock_time "soarca/test/unittest/mocks/mock_utils/time"
	"soarca/test/unittest/mocks/mock_walker"

	"soarca/pkg/cacao"
	"soarca/internal/runs/model"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestExecutePlaybook(t *testing.T) {

	playbookRepoMock := new(mocks_playbook_test.MockPlaybook)
	mockWalker := new(mock_walker.Mock_Walker)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

	newWalker := func() workflow.Walker { return mockWalker }

	executerObject := New(newWalker, playbookRepoMock, mock_reporter, mock_time)
	runId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId := "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	stepId := "step--81eff59f-d084-4324-9e0a-59e353dbd28f"

	metadata := run.Metadata{RunId: runId, PlaybookId: playbookId, StepId: stepId}

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

	mock_reporter.On("ReportStepStart", metadata, step, cacao.NewVariables(addedVariables), timeNow).Return()
	mock_reporter.On("ReportStepEnd", metadata, step, cacao.NewVariables(returnedVariables), nil, timeNow).Return()

	playbook := cacao.Playbook{ID: playbookId, PlaybookVariables: cacao.NewVariables(initialVariables)}
	playbookRepoMock.On("Get", mock.Anything, playbookId).Return(playbook, nil)
	details := workflow.Result{RunId: runId,
		PlaybookId: playbookId,
		Variables:  cacao.NewVariables(returnedVariables)}

	playbook2 := cacao.Playbook{ID: playbookId, PlaybookVariables: cacao.NewVariables(expectedVariables)}

	mockWalker.On("Execute", playbook2).Return(&details, nil)

	results, err := executerObject.Execute(metadata, step, cacao.NewVariables(addedVariables))

	mockWalker.AssertExpectations(t)
	mock_reporter.AssertExpectations(t)
	mock_time.AssertExpectations(t)
	assert.Equal(t, err, nil)
	assert.Equal(t, results, cacao.NewVariables(returnedVariables))

}
