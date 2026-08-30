package reporter_api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	api_routes "soarca/internal/transport/http/handlers"
	api_model "soarca/internal/transport/http/schema"
	runstate_model "soarca/internal/runs/state"
	mock_runstate "soarca/test/unittest/mocks/mock_runstate"
	"testing"

	runsservice "soarca/internal/runs"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestGetRunsInvocation(t *testing.T) {
	mock_runstate_reporter := &mock_runstate.MockRunState{}
	mock_runstate_reporter.On("GetRuns").Return([]runstate_model.RunEntry{}, nil)

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	api_routes.ReporterRoutesWithService(app, runsservice.New(nil, nil, mock_runstate_reporter))

	request, err := http.NewRequest("GET", "/reporter/", nil)
	if err != nil {
		t.Fail()
	}

	app.ServeHTTP(recorder, request)
	expectedString := "[]"
	assert.Equal(t, expectedString, recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)

	mock_runstate_reporter.AssertExpectations(t)
}

func TestGetRunReportInvocation(t *testing.T) {
	mock_runstate_reporter := &mock_runstate.MockRunState{}
	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	api_routes.ReporterRoutesWithService(app, runsservice.New(nil, nil, mock_runstate_reporter))

	runId0, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")

	expectedCache := `{
		"RunId":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
		"PlaybookId":"test",
		"Started":"2014-11-12T11:45:26.371Z",
		"Ended":"0001-01-01T00:00:00Z",
		"StepResults":{
		   "6ba7b810-9dad-11d1-80b4-00c04fd430c9":{
			  "RunId":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
			  "StepId":"action--test",
			  "StepRunId":"6ba7b810-9dad-11d1-80b4-00c04fd430c9",
			  "Started":"2014-11-12T11:45:26.371Z",
			  "Ended":"2014-11-12T11:45:26.371Z",
			  "Variables":{
				 "var1":{
					"type":"string",
					"name":"var1",
					"value":"testing"
				 }
			  },
			  "CommandsB64" : [],
			  "IsAutomated" : true,
			  "Status":0,
			  "Error":null
		   }
		},
		"PlaybookResult":null,
		"Status":2
	 }`
	expectedCacheData := runstate_model.RunEntry{}
	err := json.Unmarshal([]byte(expectedCache), &expectedCacheData)
	if err != nil {
		t.Log(err)
		t.Log("Could not parse data to JSON")
		t.Fail()
	}

	mock_runstate_reporter.On("GetRunReport", runId0).Return(expectedCacheData, nil)

	request, err := http.NewRequest("GET", fmt.Sprintf("/reporter/%s", runId0), nil)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	app.ServeHTTP(recorder, request)

	expectedResponse := `{
		"type":"run_status",
		"run_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
		"playbook_id":"test",
		"started":"2014-11-12T11:45:26.371Z",
		"ended":"0001-01-01T00:00:00Z",
		"status":"ongoing",
		"status_text":"this playbook is currently being executed",
		"step_results":{
		   "6ba7b810-9dad-11d1-80b4-00c04fd430c9":{
			  "run_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
			  "step_id": "action--test",
			  "step_run_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c9",
			  "started": "2014-11-12T11:45:26.371Z",
			  "ended": "2014-11-12T11:45:26.371Z",
			  "status": "successfully_executed",
			  "status_text": "step run completed successfully",
			  "Variables":{
				 "var1":{
					"type":"string",
					"name":"var1",
					"value":"testing"
				 }
			  },
			  "commands_b64" : [],
			  "automated_run" : true,
			  "executed_by" : "soarca"
		   }
		},
		"request_interval":5
	}`
	expectedResponseData := api_model.PlaybookRunReport{}
	err = json.Unmarshal([]byte(expectedResponse), &expectedResponseData)
	if err != nil {
		t.Log(err)
		t.Log("Could not parse data to JSON")
		t.Fail()
	}

	receivedData := api_model.PlaybookRunReport{}
	err = json.Unmarshal(recorder.Body.Bytes(), &receivedData)
	if err != nil {
		t.Log(err)
		t.Log("Could not parse data to JSON")
		t.Fail()
	}

	t.Log("expected response")
	t.Log(expectedResponseData)
	t.Log("received response")
	t.Log(receivedData)
	assert.Equal(t, expectedResponseData, receivedData)
	mock_runstate_reporter.AssertExpectations(t)
}
