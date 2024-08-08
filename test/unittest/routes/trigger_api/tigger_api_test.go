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

	"soarca/internal/decomposer"
	"soarca/models/cacao"
	"soarca/routes/trigger"
	mock_database_controller "soarca/test/unittest/mocks/mock_controller/database"
	mock_decomposer_controller "soarca/test/unittest/mocks/mock_controller/decomposer"
	"soarca/test/unittest/mocks/mock_decomposer"
	"soarca/test/unittest/mocks/mock_playbook_database"

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

	trigger_api := trigger.New(mock_controller, mock_database_controller)
	recorder := httptest.NewRecorder()
	trigger.Routes(app, trigger_api)

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	mock_decomposer.On("ExecuteAsync", *playbook, trigger_api.Executionsch).Return(&decomposer.ExecutionDetails{}, nil, executionId)

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
	trigger_api := trigger.New(mock_controller, mock_database_controller)
	trigger.Routes(app, trigger_api)
	mock_decomposer.On("ExecuteAsync", *playbook, trigger_api.Executionsch).Return(&decomposer.ExecutionDetails{}, nil, executionId)

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
	trigger_api := trigger.New(mock_controller, mock_database_controller)
	trigger.Routes(app, trigger_api)

	mock_decomposer.On("ExecuteAsync", *playbook, trigger_api.Executionsch).Return(&decomposer.ExecutionDetails{}, nil, executionId)

	request, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(json))
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	app.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)

	mock_decomposer.AssertExpectations(t)
}

func TestExecutionOfPlaybookByIdWithPayloadInvalidVariables(t *testing.T) {

	// ################################################################### Setup
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

	recorder_not_in_pb := httptest.NewRecorder()
	recorder_wrong_type := httptest.NewRecorder()
	recorder_not_external := httptest.NewRecorder()
	trigger_api := trigger.New(mock_controller, mock_database_controller)
	trigger.Routes(app, trigger_api)

	// ################################################################### Var not in playbook
	var_not_in_playbook := cacao.Variable{
		Name: "__not_in_playbook__",
		Type: cacao.VariableTypeString,
	}
	variables_not_in_playbook := cacao.NewVariables(var_not_in_playbook)

	json_not_in_pb, err := json.Marshal(variables_not_in_playbook)
	assert.Equal(t, err, nil)

	request_not_in_pb, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(json_not_in_pb))
	if err != nil {
		t.Fail()
	}
	app.ServeHTTP(recorder_not_in_pb, request_not_in_pb)

	// Assertions
	var result_not_in_pb map[string]interface{}
	err = json.Unmarshal(recorder_not_in_pb.Body.Bytes(), &result_not_in_pb)
	if err != nil {
		t.Fatalf("Could not unmarshal response body: %v", err)
	}
	expected_message_not_in_pb := "Cannot execute. reason: provided variables is not a valid subset of the variables for the referenced playbook [ playbook id: playbook--61a6c41e-6efc-4516-a242-dfbc5c89d562 ]"
	assert.Equal(t, 400, recorder_not_in_pb.Code)
	assert.Equal(t, expected_message_not_in_pb, result_not_in_pb["message"].(string))

	// ################################################################### Var is of wrong type
	var_wrong_type := cacao.Variable{
		Name: "__var1__",
		Type: cacao.VariableTypeInt,
	}
	variables_wrong_type := cacao.NewVariables(var_wrong_type)

	json_wrong_type, err := json.Marshal(variables_wrong_type)
	assert.Equal(t, err, nil)

	request_wrong_type, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(json_wrong_type))
	if err != nil {
		t.Fail()
	}
	app.ServeHTTP(recorder_wrong_type, request_wrong_type)
	assert.Equal(t, 400, recorder_wrong_type.Code)

	// Assertions
	var result_wrong_type map[string]interface{}
	err = json.Unmarshal(recorder_wrong_type.Body.Bytes(), &result_wrong_type)
	if err != nil {
		t.Fatalf("Could not unmarshal response body: %v", err)
	}
	expected_message_wrong_type := "Cannot execute. reason: mismatch in variables type for [ __var1__ ]: payload var type = integer, playbook var type = string"
	assert.Equal(t, 400, recorder_wrong_type.Code)
	assert.Equal(t, expected_message_wrong_type, result_wrong_type["message"].(string))

	// ################################################################### Playbook var is not external
	var_not_external := cacao.Variable{
		Name:  "__var2_not_external__",
		Type:  cacao.VariableTypeString,
		Value: "I'm not gonna be assigned :(",
	}
	variables_not_external := cacao.NewVariables(var_not_external)

	json_not_external, err := json.Marshal(variables_not_external)
	assert.Equal(t, err, nil)

	request_not_external, err := http.NewRequest("POST", "/trigger/playbook/1", bytes.NewReader(json_not_external))
	if err != nil {
		t.Fail()
	}
	app.ServeHTTP(recorder_not_external, request_not_external)
	assert.Equal(t, 400, recorder_not_external.Code)

	// Assertions
	var result_not_external map[string]interface{}
	err = json.Unmarshal(recorder_not_external.Body.Bytes(), &result_not_external)
	if err != nil {
		t.Fatalf("Could not unmarshal response body: %v", err)
	}
	expected_message_not_external := "Cannot execute. reason: playbook variable [ __var2_not_external__ ] cannot be assigned in playbook because it is not marked as external in the plabook"
	assert.Equal(t, 400, recorder_not_external.Code)
	assert.Equal(t, expected_message_not_external, result_not_external["message"].(string))

	// ################################################################### End test
	mock_decomposer.AssertExpectations(t)
}
