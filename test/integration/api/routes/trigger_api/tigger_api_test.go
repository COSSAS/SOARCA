package trigger_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	api_routes "soarca/internal/transport/http/handlers"
	trigger_handler "soarca/internal/transport/http/handlers/trigger"

	"soarca/internal/runs"
	"soarca/internal/workflow"
	"soarca/pkg/cacao"
	"soarca/internal/runs/state"
	mock_playbook_database "soarca/test/unittest/mocks/mock_playbook_database"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// testWalker reports a fixed run id instead of walking a playbook.
type testWalker struct {
	runID uuid.UUID
}

func (d *testWalker) ExecuteAsync(playbook cacao.Playbook, results chan workflow.Result) {
	if results != nil {
		results <- workflow.Result{
			RunId:      d.runID,
			PlaybookId: playbook.ID,
			Variables:  playbook.PlaybookVariables,
		}
	}
}

func (d *testWalker) Execute(playbook cacao.Playbook) (*workflow.Result, error) {
	return &workflow.Result{
		RunId:      d.runID,
		PlaybookId: playbook.ID,
		Variables:  playbook.PlaybookVariables,
	}, nil
}

type testEngine struct {
	runID uuid.UUID
}

func (e *testEngine) NewWalker() workflow.Walker {
	return &testWalker{runID: e.runID}
}

type testReports struct{}

func (r *testReports) GetRuns() ([]runstate.RunEntry, error) {
	return []runstate.RunEntry{}, nil
}

func (r *testReports) GetRunReport(runID uuid.UUID) (runstate.RunEntry, error) {
	_ = runID
	return runstate.RunEntry{}, nil
}

func close(file *os.File) {
	if err := file.Close(); err != nil {
		fmt.Println(err)
	}
}

func newTriggerHandler(runID uuid.UUID, playbookStore *mock_playbook_database.MockPlaybook) *trigger_handler.TriggerHandler {
	engine := &testEngine{runID: runID}
	runner := runs.New(engine.NewWalker, playbookStore, &testReports{})
	return trigger_handler.NewTriggerHandler(runner)
}

func TestTriggerRunOfPlaybook(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		t.Fatal(err)
	}
	defer close(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	gin.SetMode(gin.DebugMode)
	mockDatabase := new(mock_playbook_database.MockPlaybook)
	playbook := cacao.Decode(byteValue)

	runID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	mockDatabase.On("Get", "ignored", "ignored").Maybe()

	recorder := httptest.NewRecorder()
	triggerHandler := newTriggerHandler(runID, mockDatabase)
	api_routes.TriggerRoutes(app, triggerHandler)

	request, err := http.NewRequest("POST", "/trigger/playbook", bytes.NewBuffer(byteValue))
	if err != nil {
		t.Fatal(err)
	}

	app.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, `{"run_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","playbook_id":"playbook--61a6c41e-6efc-4516-a242-dfbc5c89d562"}`, recorder.Body.String())
	_ = playbook
}

func TestRunOfPlaybookById(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		t.Fatal(err)
	}
	defer close(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)

	gin.SetMode(gin.DebugMode)
	app := gin.New()
	mockDatabase := new(mock_playbook_database.MockPlaybook)
	playbook := cacao.Decode(byteValue)
	mockDatabase.On("Get", mock.Anything, "1").Return(*playbook, nil)

	runID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	recorder := httptest.NewRecorder()
	triggerHandler := newTriggerHandler(runID, mockDatabase)
	api_routes.TriggerRoutes(app, triggerHandler)

	request, err := http.NewRequest("POST", "/trigger/playbook/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	app.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestRunOfPlaybookByIdWithPayloadValidVariables(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		t.Fatal(err)
	}
	defer close(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)

	gin.SetMode(gin.DebugMode)
	app := gin.New()
	mockDatabase := new(mock_playbook_database.MockPlaybook)
	playbook := cacao.Decode(byteValue)
	mockDatabase.On("Get", mock.Anything, "1").Return(*playbook, nil)

	var1 := cacao.Variable{Name: "__var1__", Type: cacao.VariableTypeString}
	variables := cacao.NewVariables(var1)
	jsonData, err := json.Marshal(variables)
	assert.Equal(t, err, nil)

	runID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	recorder := httptest.NewRecorder()
	triggerHandler := newTriggerHandler(runID, mockDatabase)
	api_routes.TriggerRoutes(app, triggerHandler)

	request, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(jsonData))
	if err != nil {
		t.Fatal(err)
	}

	app.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestPlaybookByIdVariableNotInPlaybook(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		t.Fatal(err)
	}
	defer close(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)

	gin.SetMode(gin.DebugMode)
	app := gin.New()
	mockDatabase := new(mock_playbook_database.MockPlaybook)
	playbook := cacao.Decode(byteValue)
	mockDatabase.On("Get", mock.Anything, "1").Return(*playbook, nil)

	runID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	recorder := httptest.NewRecorder()
	triggerHandler := newTriggerHandler(runID, mockDatabase)
	api_routes.TriggerRoutes(app, triggerHandler)

	varNotInPlaybook := cacao.Variable{Name: "__not_in_playbook__", Type: cacao.VariableTypeString}
	jsonData, err := json.Marshal(cacao.NewVariables(varNotInPlaybook))
	assert.Equal(t, err, nil)

	request, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(jsonData))
	if err != nil {
		t.Fatal(err)
	}

	app.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPlaybookByIdVariableTypeMismatch(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		t.Fatal(err)
	}
	defer close(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)

	gin.SetMode(gin.DebugMode)
	app := gin.New()
	mockDatabase := new(mock_playbook_database.MockPlaybook)
	playbook := cacao.Decode(byteValue)
	mockDatabase.On("Get", mock.Anything, "1").Return(*playbook, nil)

	varWrongType := cacao.Variable{Name: "__var1__", Type: cacao.VariableTypeInt}
	jsonData, err := json.Marshal(cacao.NewVariables(varWrongType))
	assert.Equal(t, err, nil)

	runID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	recorder := httptest.NewRecorder()
	triggerHandler := newTriggerHandler(runID, mockDatabase)
	api_routes.TriggerRoutes(app, triggerHandler)

	request, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(jsonData))
	if err != nil {
		t.Fatal(err)
	}

	app.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPlaybookByIdVariableIsNotExternal(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		t.Fatal(err)
	}
	defer close(jsonFile)
	byteValue, _ := io.ReadAll(jsonFile)

	gin.SetMode(gin.DebugMode)
	app := gin.New()
	mockDatabase := new(mock_playbook_database.MockPlaybook)
	playbook := cacao.Decode(byteValue)
	mockDatabase.On("Get", mock.Anything, "1").Return(*playbook, nil)

	varNotExternal := cacao.Variable{
		Name:  "__var2_not_external__",
		Type:  cacao.VariableTypeString,
		Value: "I'm not gonna be assigned :(",
	}
	jsonData, err := json.Marshal(cacao.NewVariables(varNotExternal))
	assert.Equal(t, err, nil)

	runID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	recorder := httptest.NewRecorder()
	triggerHandler := newTriggerHandler(runID, mockDatabase)
	api_routes.TriggerRoutes(app, triggerHandler)

	request, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(jsonData))
	if err != nil {
		t.Fatal(err)
	}

	app.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
