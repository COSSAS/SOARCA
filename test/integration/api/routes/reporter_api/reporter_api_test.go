package reporter_api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	api_model "soarca/internal/transport/http/schema"
	"soarca/pkg/cacao"
	runstate_model "soarca/internal/runs/state"
	"soarca/internal/runs/model"
	"soarca/internal/reporting/reporter/downstream_reporter/runstate"
	mock_time "soarca/test/unittest/mocks/mock_utils/time"
	"testing"
	"time"

	api_routes "soarca/internal/transport/http/handlers"

	runsservice "soarca/internal/runs"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestGetRuns(t *testing.T) {
	mock_time := new(mock_time.MockTime)
	cacheReporter := runstate.New(mock_time, 10)

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
	runId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	runId1 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c1")
	runId2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c2")

	runIds := []uuid.UUID{
		runId0,
		runId1,
		runId2,
	}

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	expectedStarted, _ := time.Parse(layout, str)
	expectedEnded, _ := time.Parse(layout, "0001-01-01T00:00:00Z")

	expectedStatus := runstate_model.Ongoing.String()
	expectedStatusText, _ := api_model.GetRunStatusText(expectedStatus, "playbook")

	expectedRunsReport := []api_model.PlaybookRunReport{}
	for _, runId := range runIds {
		t.Log(runId)
		entry := api_model.PlaybookRunReport{
			Type:            "run_status",
			RunId:           runId.String(),
			PlaybookId:      "test",
			Name:            "ssh-test",
			Started:         expectedStarted,
			Ended:           expectedEnded,
			Status:          expectedStatus,
			StatusText:      expectedStatusText,
			StepResults:     map[string]api_model.StepRunReport{},
			RequestInterval: 5,
		}
		expectedRunsReport = append(expectedRunsReport, entry)
	}

	err := cacheReporter.ReportWorkflowStart(runId0, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	err = cacheReporter.ReportWorkflowStart(runId1, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(runId2, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	api_routes.ReporterRoutesWithService(app, runsservice.New(nil, nil, cacheReporter))

	request, err := http.NewRequest("GET", "/reporter/", nil)
	if err != nil {
		t.Fail()
	}

	app.ServeHTTP(recorder, request)
	expectedByte, err := json.Marshal(expectedRunsReport)
	if err != nil {
		t.Log("failed to decode expected struct to json")
		t.Fail()
	}
	expectedString := string(expectedByte)

	assert.Equal(t, expectedString, recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)

	mock_time.AssertExpectations(t)
}

func TestGetRunReport(t *testing.T) {
	// Create real runstate, create real reporter api object
	// Do runs, test retrieval via api

	mock_time := new(mock_time.MockTime)
	cacheReporter := runstate.New(mock_time, 10)

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

	runId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	runId1 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c1")
	runId2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c2")
	stepRunId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c9")
	metadata0 := run.Metadata{RunId: runId0, StepId: step1.ID, StepRunId: stepRunId0}

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	err := cacheReporter.ReportWorkflowStart(runId0, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepStart(metadata0, step1, cacao.NewVariables(expectedVariables), mock_time.Now())
	if err != nil {
		t.Fail()
	}

	err = cacheReporter.ReportWorkflowStart(runId1, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportWorkflowStart(runId2, playbook, mock_time.Now())
	if err != nil {
		t.Fail()
	}
	err = cacheReporter.ReportStepEnd(metadata0, step1, cacao.NewVariables(expectedVariables), nil, mock_time.Now())
	if err != nil {
		t.Fail()
	}

	app := gin.New()
	gin.SetMode(gin.DebugMode)

	recorder := httptest.NewRecorder()
	api_routes.ReporterRoutesWithService(app, runsservice.New(nil, nil, cacheReporter))

	expected := `{
		"type":"run_status",
		"run_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
		"playbook_id":"test",
		"name":"ssh-test",
		"started":"2014-11-12T11:45:26.371Z",
		"ended":"0001-01-01T00:00:00Z",
		"status":"ongoing",
		"status_text":"this playbook is currently being executed",
		"step_results":{
		   "6ba7b810-9dad-11d1-80b4-00c04fd430c9":{
			  "run_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
			  "step_id":"action--test",
			  "step_run_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c9",
			  "name":"ssh-tests",
			  "started":"2014-11-12T11:45:26.371Z",
			  "ended":"2014-11-12T11:45:26.371Z",
			  "status":"successfully_executed",
			  "status_text": "step run completed successfully",
			  "variables":{
				 "var1":{
					"type":"string",
					"name":"var1",
					"value":"testing"
				 }
			  },
			  "commands_b64" : ["c3NoIGxzIC1sYQ=="],
			  "automated_run" : true,
			  "executed_by" : "soarca"
		   }
		},
		"request_interval":5
	}`
	expectedData := api_model.PlaybookRunReport{}
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

	request, err := http.NewRequest("GET", fmt.Sprintf("/reporter/%s", runId0), nil)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	app.ServeHTTP(recorder, request)

	receivedData := api_model.PlaybookRunReport{}
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
