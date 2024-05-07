package reporter_api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"soarca/internal/reporter/downstream_reporter/cache"
	"soarca/models/cacao"
	"soarca/routes/reporter"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"

	mock_time "soarca/test/unittest/mocks/mock_utils/time"
)

func TestGetExecutions(t *testing.T) {
	// Create real cache, create real reporter api object
	// Do executions, test retrieval via api

	mock_time := new(mock_time.MockTime)
	cacheReporter := cache.New(mock_time)

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
	executionId0, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")
	executionId1, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c1")
	executionId2, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c2")

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	expectedExecutions := []string{
		"6ba7b810-9dad-11d1-80b4-00c04fd430c0",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c1",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c2",
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
	expectedByte, err := json.Marshal(expectedExecutions)
	if err != nil {
		t.Log("failed to decode expected struct to json")
		t.Fail()
	}
	expectedString := string(expectedByte)
	assert.Equal(t, expectedString, recorder.Body.String())
	assert.Equal(t, 200, recorder.Code)

	mock_time.AssertExpectations(t)
}

//fmt.Sprintf("/reporter/%s", executionId)
