package reporter_api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"soarca/pkg/api/reporter"
	api_model "soarca/pkg/models/api"
	cache_model "soarca/pkg/models/cache"
	mock_cache "soarca/test/unittest/mocks/mock_cache"
	"testing"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestGetExecutionsInvocation(t *testing.T) {
	mock_cache_reporter := &mock_cache.Mock_Cache{}
	mock_cache_reporter.On("GetExecutions").Return([]cache_model.ExecutionEntry{}, nil)

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	reporter.Routes(app, mock_cache_reporter)

	request, err := http.NewRequest("GET", "/reporter/", nil)
	if err != nil {
		t.Fail()
	}

	app.ServeHTTP(recorder, request)
	expectedString := "[]"
	assert.Equal(t, expectedString, recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)

	mock_cache_reporter.AssertExpectations(t)
}

func TestGetExecutionReportInvocation(t *testing.T) {
	mock_cache_reporter := &mock_cache.Mock_Cache{}
	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	reporter.Routes(app, mock_cache_reporter)

	executionId0, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")

	expectedCache := `{
		"ExecutionId":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
		"PlaybookId":"test",
		"Started":"2014-11-12T11:45:26.371Z",
		"Ended":"0001-01-01T00:00:00Z",
		"StepResults":{
		   "action--test":{
			  "ExecutionId":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
			  "StepId":"action--test",
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
	expectedCacheData := cache_model.ExecutionEntry{}
	err := json.Unmarshal([]byte(expectedCache), &expectedCacheData)
	if err != nil {
		t.Log(err)
		t.Log("Could not parse data to JSON")
		t.Fail()
	}

	mock_cache_reporter.On("GetExecutionReport", executionId0).Return(expectedCacheData, nil)

	request, err := http.NewRequest("GET", fmt.Sprintf("/reporter/%s", executionId0), nil)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	app.ServeHTTP(recorder, request)

	expectedResponse := `{
		"type":"execution_status",
		"execution_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
		"playbook_id":"test",
		"started":"2014-11-12T11:45:26.371Z",
		"ended":"0001-01-01T00:00:00Z",
		"status":"ongoing",
		"status_text":"this playbook is currently being executed",
		"step_results":{
		   "action--test":{
			  "execution_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
			  "step_id": "action--test",
			  "started": "2014-11-12T11:45:26.371Z",
			  "ended": "2014-11-12T11:45:26.371Z",
			  "status": "successfully_executed",
			  "status_text": "step execution completed successfully",
			  "Variables":{
				 "var1":{
					"type":"string",
					"name":"var1",
					"value":"testing"
				 }
			  },
			  "commands_b64" : [],
			  "automated_execution" : true,
			  "executed_by" : "soarca"
		   }
		},
		"request_interval":5
	}`
	expectedResponseData := api_model.PlaybookExecutionReport{}
	err = json.Unmarshal([]byte(expectedResponse), &expectedResponseData)
	if err != nil {
		t.Log(err)
		t.Log("Could not parse data to JSON")
		t.Fail()
	}

	receivedData := api_model.PlaybookExecutionReport{}
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
	mock_cache_reporter.AssertExpectations(t)
}
