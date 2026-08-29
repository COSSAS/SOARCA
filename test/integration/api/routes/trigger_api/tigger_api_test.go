package trigger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	api_routes "soarca/pkg/api"
	trigger_handler "soarca/pkg/api/trigger"
	triggerservice "soarca/internal/services/trigger"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/cache"
	"soarca/pkg/models/manual"
	mock_playbook_database "soarca/test/unittest/mocks/mock_playbook_database"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type testExecutionRuntime struct {
	executionID uuid.UUID
}

func (r *testExecutionRuntime) StartExecution(ctx context.Context, playbook *cacao.Playbook, variables cacao.Variables) (uuid.UUID, error) {
	_ = ctx
	_ = playbook
	_ = variables
	return r.executionID, nil
}

func (r *testExecutionRuntime) ResumeManualStep(ctx context.Context, execID uuid.UUID, stepExecID uuid.UUID, response manual.InteractionResponse) error {
	_ = ctx
	_ = execID
	_ = stepExecID
	_ = response
	return nil
}

func (r *testExecutionRuntime) GetExecutionStatus(ctx context.Context, execID uuid.UUID) (cache.ExecutionEntry, error) {
	_ = ctx
	_ = execID
	return cache.ExecutionEntry{}, nil
}

func close(file *os.File) {
	if err := file.Close(); err != nil {
		fmt.Println(err)
	}
}

func newTriggerHandler(executionID uuid.UUID, playbookStore *mock_playbook_database.MockPlaybook) *trigger_handler.TriggerHandler {
	return trigger_handler.NewTriggerHandler(triggerservice.New(&testExecutionRuntime{executionID: executionID}, playbookStore))
}

func TestTriggerExecutionOfPlaybook(t *testing.T) {
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

	executionID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	mockDatabase.On("Get", "ignored", "ignored").Maybe()

	recorder := httptest.NewRecorder()
	triggerHandler := newTriggerHandler(executionID, mockDatabase)
	api_routes.TriggerRoutes(app, triggerHandler)

	request, err := http.NewRequest("POST", "/trigger/playbook", bytes.NewBuffer(byteValue))
	if err != nil {
		t.Fatal(err)
	}

	app.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, `{"execution_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","payload":"playbook--61a6c41e-6efc-4516-a242-dfbc5c89d562"}`, recorder.Body.String())
	_ = playbook
}

func TestExecutionOfPlaybookById(t *testing.T) {
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

	executionID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	recorder := httptest.NewRecorder()
	triggerHandler := newTriggerHandler(executionID, mockDatabase)
	api_routes.TriggerRoutes(app, triggerHandler)

	request, err := http.NewRequest("POST", "/trigger/playbook/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	app.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestExecutionOfPlaybookByIdWithPayloadValidVariables(t *testing.T) {
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

	executionID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	recorder := httptest.NewRecorder()
	triggerHandler := newTriggerHandler(executionID, mockDatabase)
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

	executionID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	recorder := httptest.NewRecorder()
	triggerHandler := newTriggerHandler(executionID, mockDatabase)
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

	executionID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	recorder := httptest.NewRecorder()
	triggerHandler := newTriggerHandler(executionID, mockDatabase)
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

	executionID := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	recorder := httptest.NewRecorder()
	triggerHandler := newTriggerHandler(executionID, mockDatabase)
	api_routes.TriggerRoutes(app, triggerHandler)

	request, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(jsonData))
	if err != nil {
		t.Fatal(err)
	}

	app.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
