package reporter_api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"soarca/internal/reporter/downstream_reporter/cache"
	api_model "soarca/models/api"
	"soarca/models/cacao"
	cache_model "soarca/models/cache"
	"soarca/routes/reporter"
	mock_time "soarca/test/unittest/mocks/mock_utils/time"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestGetExecutions(t *testing.T) {

	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time, 10)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	step1 := cacao.Step{
		Type:          "action",
		ID:            "action--test",
		Name:          "ssh-tests",
		StepVariables: cacao.NewVariables(expectedVariables),
		Commands:      []cacao.Command{expectedCommand},
		Cases:         map[string]string{},
		OnCompletion:  "end--test",
		Agent:         "agent1",
		Targets:       []string{"target1"},
	}

	end := cacao.Step{
		Type: "end",
		ID:   "end--test",
		Name: "end step",
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name: "user",
		ID:   "auth1",
	}

	expectedTarget := cacao.AgentTarget{
		Name:               "sometarget",
		AuthInfoIdentifier: "auth1",
		ID:                 "target1",
	}

	expectedAgent := cacao.AgentTarget{
		Type: "soarca",
		Name: "soarca-ssh",
	}

	playbook := cacao.Playbook{
		ID:                            "test",
		Type:                          "test",
		Name:                          "ssh-test",
		WorkflowStart:                 step1.ID,
		AuthenticationInfoDefinitions: map[string]cacao.AuthenticationInformation{"id": expectedAuth},
		AgentDefinitions:              map[string]cacao.AgentTarget{"agent1": expectedAgent},
		TargetDefinitions:             map[string]cacao.AgentTarget{"target1": expectedTarget},

		Workflow: map[string]cacao.Step{step1.ID: step1, end.ID: end},
	}
	executionId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	executionId1 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c1")
	executionId2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c2")

	executionIds := []uuid.UUID{
		executionId0,
		executionId1,
		executionId2,
	}

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	expectedStarted, _ := time.Parse(layout, str)
	expectedEnded, _ := time.Parse(layout, "0001-01-01T00:00:00Z")

	expectedStatus := cache_model.Ongoing.String()
	expectedStatusText, _ := api_model.GetCacheStatusText(expectedStatus, "playbook")

	expectedExecutionsReport := []api_model.PlaybookExecutionReport{}
	for _, executionId := range executionIds {
		t.Log(executionId)
		entry := api_model.PlaybookExecutionReport{
			Type:            "execution_status",
			ExecutionId:     executionId.String(),
			PlaybookId:      "test",
			Name:            "ssh-test",
			Started:         expectedStarted,
			Ended:           expectedEnded,
			Status:          expectedStatus,
			StatusText:      expectedStatusText,
			StepResults:     map[string]api_model.StepExecutionReport{},
			RequestInterval: 5,
		}
		expectedExecutionsReport = append(expectedExecutionsReport, entry)
	}

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook)
	if err != nil {
		t.Fail()
	}

	err = cacheReporter.ReportWorkflowStart(executionId1, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId2, playbook)
	if err != nil {
		t.Fail()
	}

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	reporter.Routes(app, cacheReporter)

	request, err := http.NewRequest("GET", "/reporter/", nil)
	if err != nil {
		t.Fail()
	}

	app.ServeHTTP(recorder, request)
	expectedByte, err := json.Marshal(expectedExecutionsReport)
	if err != nil {
		t.Log("failed to decode expected struct to json")
		t.Fail()
	}
	expectedString := string(expectedByte)

	assert.Equal(t, expectedString, recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)

	mock_time.AssertExpectations(t)
}

func TestGetExecutionReport(t *testing.T) {
	// Create real cache, create real reporter api object
	// Do executions, test retrieval via api

	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time, 10)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	step1 := cacao.Step{
		Type:          "action",
		ID:            "action--test",
		Name:          "ssh-tests",
		StepVariables: cacao.NewVariables(expectedVariables),
		Commands:      []cacao.Command{expectedCommand},
		Cases:         map[string]string{},
		OnCompletion:  "end--test",
		Agent:         "agent1",
		Targets:       []string{"target1"},
	}

	end := cacao.Step{
		Type: "end",
		ID:   "end--test",
		Name: "end step",
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name: "user",
		ID:   "auth1",
	}

	expectedTarget := cacao.AgentTarget{
		Name:               "sometarget",
		AuthInfoIdentifier: "auth1",
		ID:                 "target1",
	}

	expectedAgent := cacao.AgentTarget{
		Type: "soarca",
		Name: "soarca-ssh",
	}

	playbook := cacao.Playbook{
		ID:                            "test",
		Type:                          "test",
		Name:                          "ssh-test",
		WorkflowStart:                 step1.ID,
		AuthenticationInfoDefinitions: map[string]cacao.AuthenticationInformation{"id": expectedAuth},
		AgentDefinitions:              map[string]cacao.AgentTarget{"agent1": expectedAgent},
		TargetDefinitions:             map[string]cacao.AgentTarget{"target1": expectedTarget},
		Workflow:                      map[string]cacao.Step{step1.ID: step1, end.ID: end},
	}

	executionId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	executionId1 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c1")
	executionId2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c2")

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	err := cacheReporter.ReportWorkflowStart(executionId0, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepStart(executionId0, step1, cacao.NewVariables(expectedVariables))
	if err != nil {
		t.Fail()
	}

	err = cacheReporter.ReportWorkflowStart(executionId1, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(executionId2, playbook)
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepEnd(executionId0, step1, cacao.NewVariables(expectedVariables), nil)
	if err != nil {
		t.Fail()
	}

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	reporter.Routes(app, cacheReporter)

	expected := `{
		"type":"execution_status",
		"execution_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
		"playbook_id":"test",
		"name":"ssh-test",
		"started":"2014-11-12T11:45:26.371Z",
		"ended":"0001-01-01T00:00:00Z",
		"status":"ongoing",
		"status_text":"this playbook is currently being executed",
		"step_results":{
		   "action--test":{
			  "execution_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
			  "step_id":"action--test",
			  "name":"ssh-tests",
			  "started":"2014-11-12T11:45:26.371Z",
			  "ended":"2014-11-12T11:45:26.371Z",
			  "status":"successfully_executed",
			  "status_text": "step execution completed successfully",
			  "variables":{
				 "var1":{
					"type":"string",
					"name":"var1",
					"value":"testing"
				 }
			  },
			  "commands_b64" : ["c3NoIGxzIC1sYQ=="],
			  "automated_execution" : true,
			  "executed_by" : "soarca"
		   }
		},
		"request_interval":5
	}`
	expectedData := api_model.PlaybookExecutionReport{}
	err = json.Unmarshal([]byte(expected), &expectedData)
	if err != nil {
		t.Log(err)
		t.Log("Could not parse data to JSON")
		t.Fail()
	}
	t.Log("expected")
	b, err := json.MarshalIndent(expectedData, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(b))

	request, err := http.NewRequest("GET", fmt.Sprintf("/reporter/%s", executionId0), nil)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	app.ServeHTTP(recorder, request)

	receivedData := api_model.PlaybookExecutionReport{}
	err = json.Unmarshal(recorder.Body.Bytes(), &receivedData)
	if err != nil {
		t.Log(err)
		t.Log("Could not parse data to JSON")
		t.Fail()
	}
	t.Log("received")
	t.Log(receivedData)

	assert.Equal(t, expectedData, receivedData)

	mock_time.AssertExpectations(t)
}
