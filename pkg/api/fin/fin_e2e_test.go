package fin_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	finservice "soarca/internal/services/fin"
	"soarca/internal/storage/storagetest"
	"soarca/pkg/api/fin"
	"soarca/pkg/core/capability"
	fincapability "soarca/pkg/core/capability/fin"
	"soarca/pkg/core/capability/fin/queue"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	finmodels "soarca/pkg/models/fin"
	timeUtil "soarca/pkg/utils/time"
	"soarca/test/unittest/mocks/mock_guid"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

// TestFullFinProtocolFlow is an end-to-end test of the whole Fin protocol
// stack wired together the same way internal/controller/controller.go
// wires it in production - a real fincapability.Capability enqueuing onto
// a real queue.Queue, and a real fin.FinHandler serving the actual
// register/poll/submit-result HTTP routes over that same queue - with no
// mocks standing in for either side. It exercises the full lifecycle this
// protocol exists for: register -> Execute() enqueues a job -> poll claims
// it -> submit result -> Execute() returns the result to the step machinery.
func TestFullFinProtocolFlow(t *testing.T) {
	repo := storagetest.New(t).Fins()
	jobQueue := queue.New()
	t.Cleanup(jobQueue.Close)

	fixedFinId := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	finGuidMock := new(mock_guid.Mock_Guid)
	finGuidMock.On("New").Return(fixedFinId)

	registry := finservice.NewRegistry(repo, finservice.RegistryConfig{
		RegistrationToken: "test-registration-token",
		StaleAfter:        2 * time.Minute,
	}, finGuidMock)
	workService := finservice.NewWorkService(repo, jobQueue, finservice.WorkServiceConfig{
		LongPollTimeoutSeconds: 1,
		JobLeaseSeconds:        60,
	})
	handler := fin.NewFinHandler(registry, workService, fin.Config{
		RegistrationToken:      "test-registration-token",
		PollIntervalSeconds:    5,
		LongPollTimeoutSeconds: 1,
		JobLeaseSeconds:        60,
		StaleAfter:             2 * time.Minute,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	finRoutes := router.Group("/fin")
	{
		finRoutes.POST("/register", handler.Register)
		finRoutes.GET("/", handler.List)
		finRoutes.GET(":fin_id", handler.Get)
		finRoutes.DELETE(":fin_id", handler.Delete)

		authenticated := finRoutes.Group("")
		authenticated.Use(handler.RequireFinToken)
		{
			authenticated.POST("/poll", handler.Poll)
			authenticated.PUT("jobs/:job_id", handler.SubmitResult)
			authenticated.PATCH("jobs/:job_id/status", handler.StatusPing)
			authenticated.DELETE("/", handler.Unregister)
		}
	}

	// Register a Fin declaring the capability type the step below targets.
	registerRecorder := doTestRequest(router, http.MethodPost, "/fin/register", finmodels.RegisterRequest{
		RegistrationToken: "test-registration-token",
		DisplayName:       "e2e-test-fin",
		Capabilities:      []finmodels.Capability{{Type: "custom-ssh-fin"}},
	}, "")
	if registerRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201 registering fin, got %d: %s", registerRecorder.Code, registerRecorder.Body.String())
	}
	var registered finmodels.RegisterResponse
	if err := json.Unmarshal(registerRecorder.Body.Bytes(), &registered); err != nil {
		t.Fatal(err)
	}

	// Emulate the action executor's fallback capability, the same
	// mechanism NewDecomposer() wires up in production (see
	// action.Executor.SetFinFallback).
	stepGuidMock := new(mock_guid.Mock_Guid)
	jobId := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	stepGuidMock.On("New").Return(jobId)
	finCap := fincapability.New(fincapability.Dependencies{
		Queue:      jobQueue,
		GUID:       stepGuidMock,
		Store:      repo,
		Time:       &timeUtil.Time{},
		StaleAfter: 2 * time.Minute,
	})

	executionId := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	metadata := execution.Metadata{
		ExecutionId:     executionId,
		PlaybookId:      "playbook--e2e",
		StepId:          "action--e2e",
		StepExecutionId: uuid.MustParse("55555555-5555-5555-5555-555555555555"),
	}
	commandContext := capability.Context{
		Agent:    cacao.AgentTarget{Type: "custom-ssh-fin"},
		Commands: []cacao.Command{{Type: "manual", Command: "sudo systemctl restart nginx"}},
		Step:     cacao.Step{Timeout: 5000},
	}

	// Execute blocks (bounded by the step's timeout) until a Fin claims and
	// resolves the job, so it must run concurrently with the poll/submit
	// calls below - exactly like a real action executor waiting on a real
	// external Fin process.
	type executeOutcome struct {
		variables cacao.Variables
		err       error
	}
	outcome := make(chan executeOutcome, 1)
	go func() {
		variables, err := finCap.Execute(metadata, commandContext)
		outcome <- executeOutcome{variables, err}
	}()

	// Poll claims the job Execute() just enqueued. The queue's Claim()
	// blocks internally (see queue.Queue.Claim) until notified or the
	// long-poll timeout elapses, so a single call here safely races with
	// the Execute() goroutine above regardless of goroutine scheduling.
	pollRecorder := doTestRequest(router, http.MethodPost, "/fin/poll", nil, registered.FinToken)
	if pollRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 polling for the job, got %d: %s", pollRecorder.Code, pollRecorder.Body.String())
	}
	var pollResponse finmodels.PollResponse
	if err := json.Unmarshal(pollRecorder.Body.Bytes(), &pollResponse); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, pollResponse.Job.JobId, jobId)
	assert.Equal(t, pollResponse.Job.RunId, executionId)
	assert.Equal(t, pollResponse.Job.CapabilityType, "custom-ssh-fin")
	assert.Equal(t, len(pollResponse.Job.Commands), 1)
	assert.Equal(t, pollResponse.Job.Commands[0].Command, "sudo systemctl restart nginx")

	// Submit the result, as the polling Fin would after actually running
	// the command.
	resultRecorder := doTestRequest(router, http.MethodPut, "/fin/jobs/"+jobId.String(), finmodels.ResultRequest{
		JobResult: finmodels.JobResult{
			State: finmodels.JobStateSuccess,
			Variables: cacao.NewVariables(cacao.Variable{
				Type:  "string",
				Name:  "__restarted__",
				Value: "true",
			}),
		},
	}, registered.FinToken)
	assert.Equal(t, resultRecorder.Code, http.StatusNoContent)

	select {
	case result := <-outcome:
		if result.err != nil {
			t.Fatalf("expected Execute to succeed, got error: %v", result.err)
		}
		restarted, ok := result.variables["__restarted__"]
		if !ok {
			t.Fatal("expected __restarted__ variable to be returned from the fin result")
		}
		assert.Equal(t, restarted.Value, "true")
	case <-time.After(2 * time.Second):
		t.Fatal("Execute() did not return after the fin submitted its result")
	}
}

// doTestRequest is a standalone equivalent of fin_api_test.go's (internal,
// package fin) doRequest helper - this file lives in package fin_test (an
// external test package, needed to import fincapability without an import
// cycle), so it cannot reuse that unexported helper directly.
func doTestRequest(router *gin.Engine, method string, path string, body any, bearer string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		reader = bytes.NewReader(bodyBytes)
	} else {
		reader = bytes.NewReader(nil)
	}
	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		request.Header.Set("Authorization", "Bearer "+bearer)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
