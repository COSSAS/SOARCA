package manual_api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	api_routes "soarca/internal/transport/http/handlers"
	manual_api "soarca/internal/transport/http/handlers/manual"
	"soarca/internal/workflow/capability"
	apiModel "soarca/internal/transport/http/schema"
	"soarca/pkg/cacao"
	"soarca/internal/manual/model"
	"soarca/internal/runs/model"
	"soarca/test/unittest/mocks/mock_manual_inbox_storage"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestGetPendingCommandsCalled(t *testing.T) {
	mock_manual_inbox_storage := mock_manual_inbox_storage.MockInboxStorage{}
	manualApiHandler := manual_api.NewManualHandler(&mock_manual_inbox_storage)

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	api_routes.ManualRoutes(app, manualApiHandler)

	mock_manual_inbox_storage.On("GetPendingCommands").Return([]manual.CommandInfo{}, nil)

	request, err := http.NewRequest("GET", "/manual/", nil)
	if err != nil {
		t.Fail()
	}

	app.ServeHTTP(recorder, request)
	expectedString := "[]"

	assert.Equal(t, expectedString, recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)

	mock_manual_inbox_storage.AssertExpectations(t)
}

func TestGetPendingCommandCalled(t *testing.T) {
	mock_manual_inbox_storage := mock_manual_inbox_storage.MockInboxStorage{}
	manualApiHandler := manual_api.NewManualHandler(&mock_manual_inbox_storage)

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	api_routes.ManualRoutes(app, manualApiHandler)
	testExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testStepRunId := "71a4d52c-6efc-4516-a242-dfbc5c89d999"
	path := "/manual/" + testExecId + "/" + testStepRunId
	runMetadata := run.Metadata{
		RunId: uuid.MustParse(testExecId), StepRunId: uuid.MustParse(testStepRunId),
	}

	testEmptyResponsePendingCommand := apiModel.PendingCommandData{
		Type:      "manual-command-info",
		RunId:     "00000000-0000-0000-0000-000000000000",
		StepRunId: "00000000-0000-0000-0000-000000000000",
		Commands:  []apiModel.ManualCommand{},
		Targets:   []capability.ResolvedTarget{},
	}
	emptyCommandInfoList := manual.CommandInfo{}

	mock_manual_inbox_storage.On("GetPendingCommand", runMetadata).Return(emptyCommandInfoList, nil)

	request, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Fail()
	}

	app.ServeHTTP(recorder, request)

	expectedJSON, err := json.Marshal(testEmptyResponsePendingCommand)
	if err != nil {
		t.Fatalf("Error marshalling expected JSON: %v", err)
	}
	t.Log("response:")
	t.Log(recorder.Body.String())

	expectedString := string(expectedJSON)
	assert.Equal(t, expectedString, recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)

	mock_manual_inbox_storage.AssertExpectations(t)
}

// PUT /manual/{exec_id}/{step_run_id} resolves the pending command
// identified by the path - the same resource GET identifies - rather than a
// generic POST /manual/continue carrying its own ids in the body.
func TestPutContinueCalled(t *testing.T) {
	mock_manual_inbox_storage := mock_manual_inbox_storage.MockInboxStorage{}
	manualApiHandler := manual_api.NewManualHandler(&mock_manual_inbox_storage)

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	api_routes.ManualRoutes(app, manualApiHandler)
	testExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testStepId := "61a4d52c-6efc-4516-a242-dfbc5c89d312"
	testStepRunId := "71a4d52c-6efc-4516-a242-dfbc5c89d999"
	testPlaybookId := "21a4d52c-6efc-4516-a242-dfbc5c89d312"
	path := "/manual/" + testExecId + "/" + testStepRunId

	// The metadata built from the path alone, used to look up the pending
	// command.
	lookupMetadata := run.Metadata{
		RunId:     uuid.MustParse(testExecId),
		StepRunId: uuid.MustParse(testStepRunId),
	}
	// The full metadata of the pending command as returned by the lookup;
	// this - not the (ids-free) request body - is what flows into
	// PostContinue.
	fullMetadata := run.Metadata{
		RunId:      uuid.MustParse(testExecId),
		StepId:     testStepId,
		StepRunId:  uuid.MustParse(testStepRunId),
		PlaybookId: testPlaybookId,
	}

	testManualUpdatePayload := apiModel.ManualOutArgsUpdatePayload{
		Type:           "manual-step-response",
		ResponseStatus: "success",
		ResponseOutArgs: cacao.Variables{
			"testvar": {
				Type:  "string",
				Name:  "testvar",
				Value: "testing!",
			},
		},
	}

	pendingCommand := manual.CommandInfo{Metadata: fullMetadata}

	testManualResponse := manual.Response{
		Metadata:       fullMetadata,
		ResponseStatus: "success",
		OutArgsVariables: cacao.Variables{
			"testvar": {
				Type:  "string",
				Name:  "testvar",
				Value: "testing!",
			},
		},
	}
	jsonData, err := json.Marshal(testManualUpdatePayload)
	if err != nil {
		t.Fatalf("Error marshalling JSON: %v", err)
	}

	mock_manual_inbox_storage.On("GetPendingCommand", lookupMetadata).Return(pendingCommand, nil)
	mock_manual_inbox_storage.On("PostContinue", testManualResponse).Return(nil)

	request, err := http.NewRequest("PUT", path, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fail()
	}

	app.ServeHTTP(recorder, request)
	t.Log(recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)

	mock_manual_inbox_storage.AssertExpectations(t)
}

func TestPutContinueFailsOnNonMatchingOutArgNames(t *testing.T) {
	mock_manual_inbox_storage := mock_manual_inbox_storage.MockInboxStorage{}
	manualApiHandler := manual_api.NewManualHandler(&mock_manual_inbox_storage)

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	api_routes.ManualRoutes(app, manualApiHandler)
	testExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testStepRunId := "71a4d52c-6efc-4516-a242-dfbc5c89d999"
	path := "/manual/" + testExecId + "/" + testStepRunId

	testManualUpdatePayload := apiModel.ManualOutArgsUpdatePayload{
		Type:           "manual-step-response",
		ResponseStatus: "success",
		ResponseOutArgs: cacao.Variables{
			"__this_var__": {
				Type:  "string",
				Name:  "__is_invalid__",
				Value: "testing!",
			},
		},
	}

	manualUpdatePayloadJson, err := json.Marshal(testManualUpdatePayload)
	if err != nil {
		t.Fatalf("Error marshalling JSON: %v", err)
	}

	request, err := http.NewRequest("PUT", path, bytes.NewBuffer(manualUpdatePayloadJson))
	if err != nil {
		t.Fail()
	}

	expectedErr := errors.New("variable name mismatch for variable __this_var__: has different name property: __is_invalid__")

	app.ServeHTTP(recorder, request)
	t.Log(recorder.Body.String())
	assert.Equal(t, 400, recorder.Code)
	mock_manual_inbox_storage.AssertExpectations(t)
	assert.Equal(t, true, strings.Contains(recorder.Body.String(), expectedErr.Error()))
}
