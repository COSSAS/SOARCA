package reporter_api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	api_model "soarca/models/api"
	cache_model "soarca/models/cache"
	"soarca/routes/reporter"
	mock_ds_reporter "soarca/test/unittest/mocks/mock_reporter"
	"testing"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestGetExecutionsInvocation(t *testing.T) {
	mock_cache_reporter := &mock_ds_reporter.Mock_Downstream_Reporter{}
	mock_cache_reporter.On("GetExecutionsIds").Return([]string{})

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
	mock_cache_reporter := &mock_ds_reporter.Mock_Downstream_Reporter{}
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
		"Type":"execution_status",
		"ExecutionId":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
		"PlaybookId":"test",
		"Started":"2014-11-12 11:45:26.371 +0000 UTC",
		"Ended":"0001-01-01 00:00:00 +0000 UTC",
		"Status":"ongoing",
		"StatusText":"",
		"StepResults":{
		   "action--test":{
			  "ExecutionId":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
			  "StepId":"action--test",
			  "Started":"2014-11-12 11:45:26.371 +0000 UTC",
			  "Ended":"2014-11-12 11:45:26.371 +0000 UTC",
			  "Status":"successfully_executed",
			  "StatusText":"",
			  "Error":"",
			  "Variables":{
				 "var1":{
					"type":"string",
					"name":"var1",
					"value":"testing"
				 }
			  }
		   }
		},
		"Error":"",
		"RequestInterval":5
	}`
	expectedResponseData := api_model.PlaybookExecutionReport{}
	err = json.Unmarshal([]byte(expectedResponse), &expectedResponseData)
	if err != nil {
		t.Log(err)
		t.Log("Could not parse data to JSON")
		t.Fail()
	}

	receivedData := api_model.PlaybookExecutionReport{}
	err = json.Unmarshal([]byte(recorder.Body.String()), &receivedData)
	if err != nil {
		t.Log(err)
		t.Log("Could not parse data to JSON")
		t.Fail()
	}

	assert.Equal(t, expectedResponseData, receivedData)
	mock_cache_reporter.AssertExpectations(t)
}
