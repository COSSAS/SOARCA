package trigger_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"soarca/pkg/core/decomposer"
	"soarca/pkg/models/cacao"
	"soarca/test/unittest/mocks/mock_decomposer"
	"soarca/test/unittest/mocks/mock_playbook_database"
	"testing"

	api_routes "soarca/pkg/api"

	trigger_handler "soarca/pkg/api/trigger"
	mock_database_controller "soarca/test/unittest/mocks/mock_controller/database"
	mock_decomposer_controller "soarca/test/unittest/mocks/mock_controller/decomposer"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestTriggerExecutionOfPlaybook(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	app := gin.New()
	gin.SetMode(gin.DebugMode)
	mock_decomposer := new(mock_decomposer.Mock_Decomposer)
	mock_controller := new(mock_decomposer_controller.Mock_Controller)
	mock_database_controller := new(mock_database_controller.Mock_Controller)
	mock_controller.On("NewDecomposer").Return(mock_decomposer)
	playbook := cacao.Decode(byteValue)

	recorder := httptest.NewRecorder()
	triggerHandler := trigger_handler.NewTriggerHandler(mock_controller, mock_database_controller)
	api_routes.TriggerRoutes(app, triggerHandler)
	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	mock_decomposer.On("ExecuteAsync", *playbook, triggerHandler.ExecutionsChannel).Return(&decomposer.ExecutionDetails{}, nil, executionId)

	request, err := http.NewRequest("POST", "/trigger/playbook", bytes.NewBuffer(byteValue))
	if err != nil {
		t.Fail()
	}

	expected_return_string := `{"execution_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","payload":"playbook--61a6c41e-6efc-4516-a242-dfbc5c89d562"}`
	app.ServeHTTP(recorder, request)
	assert.Equal(t, expected_return_string, recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)
	mock_decomposer.AssertExpectations(t)
}

func TestExecutionOfPlaybookById(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	gin.SetMode(gin.DebugMode)
	app := gin.New()
	mock_decomposer := new(mock_decomposer.Mock_Decomposer)
	mock_controller := new(mock_decomposer_controller.Mock_Controller)
	mock_database := new(mock_playbook_database.MockPlaybook)
	mock_database_controller := new(mock_database_controller.Mock_Controller)
	mock_database_controller.On("GetDatabaseInstance").Return(mock_database)
	playbook := cacao.Decode(byteValue)
	mock_database.On("Read", "1").Return(*playbook, nil)
	mock_controller.On("NewDecomposer").Return(mock_decomposer)
	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	recorder := httptest.NewRecorder()
	triggerHandler := trigger_handler.NewTriggerHandler(mock_controller, mock_database_controller)
	api_routes.TriggerRoutes(app, triggerHandler)
	mock_decomposer.On("ExecuteAsync", *playbook, triggerHandler.ExecutionsChannel).Return(&decomposer.ExecutionDetails{}, nil, executionId)

	request, err := http.NewRequest("POST", "/trigger/playbook/1", nil)
	if err != nil {
		t.Fail()
	}
	app.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	mock_decomposer.AssertExpectations(t)
}

func TestExecutionOfPlaybookByIdWithPayloadValidVariables(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	gin.SetMode(gin.DebugMode)
	app := gin.New()

	mock_decomposer := new(mock_decomposer.Mock_Decomposer)
	mock_controller := new(mock_decomposer_controller.Mock_Controller)

	mock_database := new(mock_playbook_database.MockPlaybook)
	mock_database_controller := new(mock_database_controller.Mock_Controller)
	mock_database_controller.On("GetDatabaseInstance").Return(mock_database)

	playbook := cacao.Decode(byteValue)

	mock_database.On("Read", "1").Return(*playbook, nil)
	mock_controller.On("NewDecomposer").Return(mock_decomposer)
	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	var1 := cacao.Variable{
		Name: "__var1__",
		Type: cacao.VariableTypeString,
	}
	variables := cacao.NewVariables(var1)

	json, err := json.Marshal(variables)
	assert.Equal(t, err, nil)

	recorder := httptest.NewRecorder()
	triggerHandler := trigger_handler.NewTriggerHandler(mock_controller, mock_database_controller)
	api_routes.TriggerRoutes(app, triggerHandler)

	mock_decomposer.On("ExecuteAsync", *playbook, triggerHandler.ExecutionsChannel).Return(&decomposer.ExecutionDetails{}, nil, executionId)

	request, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(json))
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	app.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)

	mock_decomposer.AssertExpectations(t)
}

func TestPlaybookByIdVariableNotInPlaybook(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	gin.SetMode(gin.DebugMode)
	app := gin.New()
	mock_decomposer := new(mock_decomposer.Mock_Decomposer)
	mock_controller := new(mock_decomposer_controller.Mock_Controller)
	mock_database := new(mock_playbook_database.MockPlaybook)
	mock_database_controller := new(mock_database_controller.Mock_Controller)
	mock_database_controller.On("GetDatabaseInstance").Return(mock_database)
	playbook := cacao.Decode(byteValue)
	mock_database.On("Read", "1").Return(*playbook, nil)
	mock_controller.On("NewDecomposer").Return(mock_decomposer)

	recorder := httptest.NewRecorder()
	triggerHandler := trigger_handler.NewTriggerHandler(mock_controller, mock_database_controller)
	api_routes.TriggerRoutes(app, triggerHandler)

	var_not_in_playbook := cacao.Variable{
		Name: "__not_in_playbook__",
		Type: cacao.VariableTypeString,
	}
	variablesNotInPlaybook := cacao.NewVariables(var_not_in_playbook)

	jsonNotInPlaybook, err := json.Marshal(variablesNotInPlaybook)
	assert.Equal(t, err, nil)

	requestNotInPlaybook, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(jsonNotInPlaybook))
	if err != nil {
		t.Fail()
	}
	app.ServeHTTP(recorder, requestNotInPlaybook)

	// Assertions
	var resultNotInPlaybook map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &resultNotInPlaybook)
	if err != nil {
		t.Fatalf("Could not unmarshal response body: %v", err)
	}
	notInPlaybookError := "Cannot execute. reason: provided variables is not a valid subset of the variables for the referenced playbook [ playbook id: playbook--61a6c41e-6efc-4516-a242-dfbc5c89d562 ]"
	assert.Equal(t, 400, recorder.Code)
	assert.Equal(t, notInPlaybookError, resultNotInPlaybook["message"].(string))
}

func TestPlaybookByIdVariableTypeMismatch(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	gin.SetMode(gin.DebugMode)
	app := gin.New()
	mock_decomposer := new(mock_decomposer.Mock_Decomposer)
	mock_controller := new(mock_decomposer_controller.Mock_Controller)
	mock_database := new(mock_playbook_database.MockPlaybook)
	mock_database_controller := new(mock_database_controller.Mock_Controller)
	mock_database_controller.On("GetDatabaseInstance").Return(mock_database)
	playbook := cacao.Decode(byteValue)
	mock_database.On("Read", "1").Return(*playbook, nil)
	mock_controller.On("NewDecomposer").Return(mock_decomposer)

	recorder := httptest.NewRecorder()
	triggerHandler := trigger_handler.NewTriggerHandler(mock_controller, mock_database_controller)
	api_routes.TriggerRoutes(app, triggerHandler)

	var_wrong_type := cacao.Variable{
		Name: "__var1__",
		Type: cacao.VariableTypeInt,
	}
	variablesWrongType := cacao.NewVariables(var_wrong_type)

	jsonWrongType, err := json.Marshal(variablesWrongType)
	assert.Equal(t, err, nil)

	requestWrongType, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(jsonWrongType))
	if err != nil {
		t.Fail()
	}
	app.ServeHTTP(recorder, requestWrongType)
	assert.Equal(t, 400, recorder.Code)

	// Assertions
	var resultWrongType map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &resultWrongType)
	if err != nil {
		t.Fatalf("Could not unmarshal response body: %v", err)
	}
	expected_message_wrong_type := "Cannot execute. reason: mismatch in variables type for [ __var1__ ]: payload var type = integer, playbook var type = string"
	assert.Equal(t, 400, recorder.Code)
	assert.Equal(t, expected_message_wrong_type, resultWrongType["message"].(string))
}

func TestPlaybookByIdVariableIsNotExternal(t *testing.T) {
	jsonFile, err := os.Open("../playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	gin.SetMode(gin.DebugMode)
	app := gin.New()
	mock_decomposer := new(mock_decomposer.Mock_Decomposer)
	mock_controller := new(mock_decomposer_controller.Mock_Controller)
	mock_database := new(mock_playbook_database.MockPlaybook)
	mock_database_controller := new(mock_database_controller.Mock_Controller)
	mock_database_controller.On("GetDatabaseInstance").Return(mock_database)
	playbook := cacao.Decode(byteValue)
	mock_database.On("Read", "1").Return(*playbook, nil)
	mock_controller.On("NewDecomposer").Return(mock_decomposer)

	recorder := httptest.NewRecorder()
	triggerHandler := trigger_handler.NewTriggerHandler(mock_controller, mock_database_controller)
	api_routes.TriggerRoutes(app, triggerHandler)

	varNotExternal := cacao.Variable{
		Name:  "__var2_not_external__",
		Type:  cacao.VariableTypeString,
		Value: "I'm not gonna be assigned :(",
	}
	variablesNotExternal := cacao.NewVariables(varNotExternal)

	jsonNotExternal, err := json.Marshal(variablesNotExternal)
	assert.Equal(t, err, nil)

	request_not_external, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(jsonNotExternal))
	if err != nil {
		t.Fail()
	}
	app.ServeHTTP(recorder, request_not_external)
	assert.Equal(t, 400, recorder.Code)

	// Assertions
	var resultNotExternal map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &resultNotExternal)
	if err != nil {
		t.Fatalf("Could not unmarshal response body: %v", err)
	}
	expectedError := "Cannot execute. reason: playbook variable [ __var2_not_external__ ] cannot be assigned in playbook because it is not marked as external in the plabook"
	assert.Equal(t, 400, recorder.Code)
	assert.Equal(t, expectedError, resultNotExternal["message"].(string))

	mock_decomposer.AssertExpectations(t)
}
