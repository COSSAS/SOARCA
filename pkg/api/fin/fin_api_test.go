package fin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"soarca/internal/database/finmemory"
	"soarca/pkg/core/capability/fin/queue"
	"soarca/pkg/models/fin"
	"soarca/test/unittest/mocks/mock_guid"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

const registrationToken = "test-registration-token"

func newTestHandler(t *testing.T) (*FinHandler, *finmemory.InMemoryFinRepository, *queue.Queue) {
	t.Helper()
	repo := finmemory.New()
	jobQueue := queue.New()
	t.Cleanup(jobQueue.Close)

	fixedId := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	guidMock := new(mock_guid.Mock_Guid)
	guidMock.On("New").Return(fixedId)

	handler := NewFinHandler(repo, jobQueue, Config{
		RegistrationToken:      registrationToken,
		PollIntervalSeconds:    5,
		LongPollTimeoutSeconds: 1,
		JobLeaseSeconds:        60,
	}, guidMock)
	return handler, repo, jobQueue
}

func newTestRouter(handler *FinHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Mirrors pkg/api.FinPublicRoutes + pkg/api.FinAdminRoutes: in
	// production these are two separate route-registration functions
	// (register/poll/jobs/status/unregister vs list/get), registered at two
	// different points relative to the global admin auth middleware, so
	// that Fin-token-authenticated calls are never also gated behind a
	// soarca_admin JWT. That split doesn't affect this test harness (which
	// has no admin auth middleware at all), so it's flattened here for
	// convenience.
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
	return router
}

func doRequest(router *gin.Engine, method string, path string, body any, bearer string) *httptest.ResponseRecorder {
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

func registerTestFin(t *testing.T, router *gin.Engine, capabilityType string) fin.RegisterResponse {
	t.Helper()
	recorder := doRequest(router, http.MethodPost, "/fin/register", fin.RegisterRequest{
		RegistrationToken: registrationToken,
		DisplayName:       "Test Fin",
		Capabilities:      []fin.Capability{{Type: capabilityType}},
	}, "")
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var response fin.RegisterResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	return response
}

func TestRegisterSucceeds(t *testing.T) {
	handler, _, _ := newTestHandler(t)
	router := newTestRouter(handler)

	response := registerTestFin(t, router, "pong")
	assert.NotEqual(t, response.FinId, "")
	assert.NotEqual(t, response.FinToken, "")
	assert.Equal(t, response.PollIntervalSeconds, 5)
}

func TestRegisterFailsWithWrongToken(t *testing.T) {
	handler, _, _ := newTestHandler(t)
	router := newTestRouter(handler)

	recorder := doRequest(router, http.MethodPost, "/fin/register", fin.RegisterRequest{
		RegistrationToken: "wrong-token",
		Capabilities:      []fin.Capability{{Type: "pong"}},
	}, "")
	assert.Equal(t, recorder.Code, http.StatusForbidden)
}

func TestRegisterFailsWithoutCapabilities(t *testing.T) {
	handler, _, _ := newTestHandler(t)
	router := newTestRouter(handler)

	recorder := doRequest(router, http.MethodPost, "/fin/register", fin.RegisterRequest{
		RegistrationToken: registrationToken,
		Capabilities:      []fin.Capability{},
	}, "")
	assert.Equal(t, recorder.Code, http.StatusBadRequest)
}

func TestRegisterFailsWhenNotConfigured(t *testing.T) {
	repo := finmemory.New()
	jobQueue := queue.New()
	defer jobQueue.Close()
	guidMock := new(mock_guid.Mock_Guid)
	handler := NewFinHandler(repo, jobQueue, Config{}, guidMock)
	router := newTestRouter(handler)

	recorder := doRequest(router, http.MethodPost, "/fin/register", fin.RegisterRequest{
		RegistrationToken: "",
		Capabilities:      []fin.Capability{{Type: "pong"}},
	}, "")
	assert.Equal(t, recorder.Code, http.StatusServiceUnavailable)
}

func TestPollRequiresFinToken(t *testing.T) {
	handler, _, _ := newTestHandler(t)
	router := newTestRouter(handler)

	recorder := doRequest(router, http.MethodPost, "/fin/poll", nil, "")
	assert.Equal(t, recorder.Code, http.StatusUnauthorized)

	recorder = doRequest(router, http.MethodPost, "/fin/poll", nil, "not-a-real-token")
	assert.Equal(t, recorder.Code, http.StatusUnauthorized)
}

func TestPollReturnsNoContentWhenNoJobIsAvailable(t *testing.T) {
	handler, _, _ := newTestHandler(t)
	router := newTestRouter(handler)

	registered := registerTestFin(t, router, "pong")

	recorder := doRequest(router, http.MethodPost, "/fin/poll", nil, registered.FinToken)
	assert.Equal(t, recorder.Code, http.StatusNoContent)
}

func TestPollReturnsEnqueuedJobAndUpdatesLastSeen(t *testing.T) {
	handler, repo, jobQueue := newTestHandler(t)
	router := newTestRouter(handler)

	registered := registerTestFin(t, router, "pong")
	before, err := repo.Get(registered.FinId)
	if err != nil {
		t.Fatal(err)
	}

	job := fin.Job{
		JobId:                 uuid.New(),
		CapabilityType:        "pong",
		LeaseExpiresInSeconds: 60,
	}
	go func() {
		_, _ = jobQueue.Enqueue(context.Background(), job)
	}()
	time.Sleep(20 * time.Millisecond)

	recorder := doRequest(router, http.MethodPost, "/fin/poll", nil, registered.FinToken)
	assert.Equal(t, recorder.Code, http.StatusOK)

	var response fin.PollResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, response.Job.JobId, job.JobId)

	after, err := repo.Get(registered.FinId)
	if err != nil {
		t.Fatal(err)
	}
	if !after.LastSeen.After(before.LastSeen) {
		t.Fatal("expected LastSeen to be updated by Poll")
	}

	// Clean up: submit a result so the Enqueue goroutine doesn't leak past
	// the test.
	_ = jobQueue.Submit(job.JobId, registered.FinId, fin.JobResult{State: fin.JobStateSuccess})
}

func TestSubmitResultRoundTrip(t *testing.T) {
	handler, repo, jobQueue := newTestHandler(t)
	router := newTestRouter(handler)

	registered := registerTestFin(t, router, "pong")

	job := fin.Job{
		JobId:                 uuid.New(),
		CapabilityType:        "pong",
		LeaseExpiresInSeconds: 60,
	}
	resultCh := make(chan fin.JobResult, 1)
	go func() {
		result, _ := jobQueue.Enqueue(context.Background(), job)
		resultCh <- result
	}()

	recorder := doRequest(router, http.MethodPost, "/fin/poll", nil, registered.FinToken)
	assert.Equal(t, recorder.Code, http.StatusOK)

	lastSeenAfterPoll, err := repo.Get(registered.FinId)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)

	recorder = doRequest(router, http.MethodPut, "/fin/jobs/"+job.JobId.String(),
		fin.ResultRequest{JobResult: fin.JobResult{State: fin.JobStateSuccess}}, registered.FinToken)
	assert.Equal(t, recorder.Code, http.StatusNoContent)

	select {
	case result := <-resultCh:
		assert.Equal(t, result.State, fin.JobStateSuccess)
	case <-time.After(time.Second):
		t.Fatal("expected the enqueued job to receive its result")
	}

	afterSubmit, err := repo.Get(registered.FinId)
	if err != nil {
		t.Fatal(err)
	}
	if !afterSubmit.LastSeen.After(lastSeenAfterPoll.LastSeen) {
		t.Fatalf("expected submitting a job result to advance LastSeen: before=%v after=%v",
			lastSeenAfterPoll.LastSeen, afterSubmit.LastSeen)
	}
}

func TestSubmitResultFailsForUnknownJob(t *testing.T) {
	handler, _, _ := newTestHandler(t)
	router := newTestRouter(handler)

	registered := registerTestFin(t, router, "pong")

	recorder := doRequest(router, http.MethodPut, "/fin/jobs/"+uuid.New().String(),
		fin.ResultRequest{JobResult: fin.JobResult{State: fin.JobStateSuccess}}, registered.FinToken)
	assert.Equal(t, recorder.Code, http.StatusNotFound)
}

func TestSubmitResultFailsWhenLeasedToAnotherFin(t *testing.T) {
	handler, _, jobQueue := newTestHandler(t)
	router := newTestRouter(handler)

	registeredA := registerTestFin(t, router, "pong")

	// A second fin registered under the same capability type, to claim the
	// job first without being the one submitting the result.
	guidMock := handler.guid.(*mock_guid.Mock_Guid)
	guidMock.ExpectedCalls = nil
	guidMock.On("New").Return(uuid.MustParse("22222222-2222-2222-2222-222222222222"))
	registeredB := registerTestFin(t, router, "pong")

	job := fin.Job{JobId: uuid.New(), CapabilityType: "pong", LeaseExpiresInSeconds: 60}
	go func() { _, _ = jobQueue.Enqueue(context.Background(), job) }()

	recorder := doRequest(router, http.MethodPost, "/fin/poll", nil, registeredB.FinToken)
	assert.Equal(t, recorder.Code, http.StatusOK)

	recorder = doRequest(router, http.MethodPut, "/fin/jobs/"+job.JobId.String(),
		fin.ResultRequest{JobResult: fin.JobResult{State: fin.JobStateSuccess}}, registeredA.FinToken)
	assert.Equal(t, recorder.Code, http.StatusForbidden)
}

func TestStatusPingExtendsLease(t *testing.T) {
	handler, repo, jobQueue := newTestHandler(t)
	router := newTestRouter(handler)

	registered := registerTestFin(t, router, "pong")

	job := fin.Job{JobId: uuid.New(), CapabilityType: "pong", LeaseExpiresInSeconds: 60}
	go func() { _, _ = jobQueue.Enqueue(context.Background(), job) }()

	recorder := doRequest(router, http.MethodPost, "/fin/poll", nil, registered.FinToken)
	assert.Equal(t, recorder.Code, http.StatusOK)

	lastSeenAfterPoll, err := repo.Get(registered.FinId)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)

	recorder = doRequest(router, http.MethodPatch, "/fin/jobs/"+job.JobId.String()+"/status",
		fin.StatusPingRequest{Progress: "running"}, registered.FinToken)
	assert.Equal(t, recorder.Code, http.StatusOK)

	var response fin.StatusPingResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, response.Action, "")

	afterPing, err := repo.Get(registered.FinId)
	if err != nil {
		t.Fatal(err)
	}
	if !afterPing.LastSeen.After(lastSeenAfterPoll.LastSeen) {
		t.Fatalf("expected a status ping to advance LastSeen: before=%v after=%v",
			lastSeenAfterPoll.LastSeen, afterPing.LastSeen)
	}

	_ = jobQueue.Submit(job.JobId, registered.FinId, fin.JobResult{State: fin.JobStateSuccess})
}

func TestUnregisterOwnRegistrationSucceeds(t *testing.T) {
	handler, repo, _ := newTestHandler(t)
	router := newTestRouter(handler)

	registered := registerTestFin(t, router, "pong")

	recorder := doRequest(router, http.MethodDelete, "/fin/", nil, registered.FinToken)
	assert.Equal(t, recorder.Code, http.StatusNoContent)

	_, err := repo.Get(registered.FinId)
	if err == nil {
		t.Fatal("expected fin to be unregistered")
	}
}

func TestAdminDeleteRemovesAnyFinsRegistration(t *testing.T) {
	handler, repo, _ := newTestHandler(t)
	router := newTestRouter(handler)

	registered := registerTestFin(t, router, "pong")

	// Admin delete is not fin-token gated at all - no Authorization header.
	recorder := doRequest(router, http.MethodDelete, "/fin/"+registered.FinId, nil, "")
	assert.Equal(t, recorder.Code, http.StatusNoContent)

	_, err := repo.Get(registered.FinId)
	if err == nil {
		t.Fatal("expected fin to be unregistered")
	}
}

func TestListAndGet(t *testing.T) {
	handler, _, _ := newTestHandler(t)
	router := newTestRouter(handler)

	registered := registerTestFin(t, router, "pong")

	recorder := doRequest(router, http.MethodGet, "/fin/", nil, "")
	assert.Equal(t, recorder.Code, http.StatusOK)
	var listResponse fin.ListResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &listResponse); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(listResponse.Fins), 1)
	assert.Equal(t, listResponse.Fins[0].Stale, false)

	recorder = doRequest(router, http.MethodGet, "/fin/"+registered.FinId, nil, "")
	assert.Equal(t, recorder.Code, http.StatusOK)
	var record fin.Record
	if err := json.Unmarshal(recorder.Body.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record.Stale, false)

	recorder = doRequest(router, http.MethodGet, "/fin/does-not-exist", nil, "")
	assert.Equal(t, recorder.Code, http.StatusNotFound)
}

func TestListAndGetMarkFinStaleAfterThreshold(t *testing.T) {
	handler, repo, _ := newTestHandler(t)
	handler.config.StaleAfterSeconds = 1
	router := newTestRouter(handler)

	registered := registerTestFin(t, router, "pong")
	if err := repo.Touch(registered.FinId, time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	recorder := doRequest(router, http.MethodGet, "/fin/"+registered.FinId, nil, "")
	assert.Equal(t, recorder.Code, http.StatusOK)
	var record fin.Record
	if err := json.Unmarshal(recorder.Body.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record.Stale, true)

	recorder = doRequest(router, http.MethodGet, "/fin/", nil, "")
	assert.Equal(t, recorder.Code, http.StatusOK)
	var listResponse fin.ListResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &listResponse); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(listResponse.Fins), 1)
	assert.Equal(t, listResponse.Fins[0].Stale, true)
}
