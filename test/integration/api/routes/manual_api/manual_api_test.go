package manual_api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	api_routes "soarca/pkg/api"
	manual_api "soarca/pkg/api/manual"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"
	manual_model "soarca/pkg/models/manual"
	"soarca/test/unittest/mocks/mock_interaction_storage"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestGetPendingCommandsCalled(t *testing.T) {
	mock_interaction_storage := mock_interaction_storage.MockInteractionStorage{}
	manualApiHandler := manual_api.NewManualHandler(&mock_interaction_storage)

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	api_routes.ManualRoutes(app, manualApiHandler)

	mock_interaction_storage.On("GetPendingCommands").Return([]manual_model.InteractionCommandData{}, 200, nil)

	request, err := http.NewRequest("GET", "/manual/", nil)
	if err != nil {
		t.Fail()
	}

	app.ServeHTTP(recorder, request)
	expectedString := "[]"

	assert.Equal(t, expectedString, recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)

	mock_interaction_storage.AssertExpectations(t)
}

func TestGetPendingCommandCalled(t *testing.T) {
	mock_interaction_storage := mock_interaction_storage.MockInteractionStorage{}
	manualApiHandler := manual_api.NewManualHandler(&mock_interaction_storage)

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	api_routes.ManualRoutes(app, manualApiHandler)
	testExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testStepId := "61a4d52c-6efc-4516-a242-dfbc5c89d312"
	path := "/manual/" + testExecId + "/" + testStepId
	executionMetadata := execution.Metadata{
		ExecutionId: uuid.MustParse(testExecId), StepId: testStepId,
	}

	testEmptyResponsePendingCommand := manual.InteractionCommandData{}

	mock_interaction_storage.On("GetPendingCommand", executionMetadata).Return(testEmptyResponsePendingCommand, 200, nil)

	request, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Fail()
	}

	app.ServeHTTP(recorder, request)
	expectedData := testEmptyResponsePendingCommand
	expectedJSON, err := json.Marshal(expectedData)
	if err != nil {
		t.Fatalf("Error marshalling expected JSON: %v", err)
	}
	expectedString := string(expectedJSON)
	assert.Equal(t, expectedString, recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)

	mock_interaction_storage.AssertExpectations(t)
}

func TestPatchContinueCalled(t *testing.T) {
	mock_interaction_storage := mock_interaction_storage.MockInteractionStorage{}
	manualApiHandler := manual_api.NewManualHandler(&mock_interaction_storage)

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	api_routes.ManualRoutes(app, manualApiHandler)
	testExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testStepId := "61a4d52c-6efc-4516-a242-dfbc5c89d312"
	testPlaybookId := "21a4d52c-6efc-4516-a242-dfbc5c89d312"
	path := "/manual/" + testExecId + "/" + testStepId

	testManualResponse := manual_model.ManualOutArgsUpdatePayload{
		Type:           "manual-step-response",
		ExecutionId:    testExecId,
		StepId:         testStepId,
		PlaybookId:     testPlaybookId,
		ResponseStatus: true,
		ResponseOutArgs: manual_model.ManualOutArgs{
			"testvar": {
				Type:  "string",
				Name:  "testvar",
				Value: "testing!",
			},
		},
	}
	jsonData, err := json.Marshal(testManualResponse)
	if err != nil {
		t.Fatalf("Error marshalling JSON: %v", err)
	}

	mock_interaction_storage.On("PostContinue", testManualResponse).Return(200, nil)

	request, err := http.NewRequest("PATCH", path, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fail()
	}

	app.ServeHTTP(recorder, request)
	t.Log(recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)

	mock_interaction_storage.AssertExpectations(t)
}
