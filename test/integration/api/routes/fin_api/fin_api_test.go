package fin_api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	finservice "soarca/internal/fins"
	storagetest "soarca/internal/store/storagetest"
	api_routes "soarca/internal/transport/http/handlers"
	fin_handler "soarca/internal/transport/http/handlers/fin"
	"soarca/internal/workflow/capability/fin/queue"
	"soarca/pkg/fins/protocol"
	"soarca/pkg/utils/guid"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const registrationToken = "test-registration-token"

// newFinTestApp wires the real registry and work service over in-memory storage,
// so these tests cover the actual route/middleware/service composition.
func newFinTestApp(t *testing.T, regToken string) *gin.Engine {
	t.Helper()

	store := storagetest.New(t)
	q := queue.New()
	t.Cleanup(q.Close)

	registry := finservice.NewRegistry(
		store.Fins(),
		finservice.RegistryConfig{RegistrationToken: regToken, StaleAfter: 10 * time.Second},
		new(guid.Guid),
	)
	workService := finservice.NewWorkService(
		store.Fins(),
		q,
		// Long poll timeout kept at 1s so "no work available" returns quickly.
		finservice.WorkServiceConfig{LongPollTimeoutSeconds: 1, JobLeaseSeconds: 60},
	)

	handler := fin_handler.NewFinHandler(registry, workService, fin_handler.Config{
		RegistrationToken:      regToken,
		PollIntervalSeconds:    5,
		LongPollTimeoutSeconds: 1,
		JobLeaseSeconds:        60,
		StaleAfter:             10 * time.Second,
	})

	gin.SetMode(gin.TestMode)
	app := gin.New()
	api_routes.FinPublic(app, handler)
	api_routes.FinAdmin(app, handler)
	return app
}

func doRequest(app *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != nil {
		encoded, _ := json.Marshal(body)
		reader = bytes.NewReader(encoded)
	} else {
		reader = bytes.NewReader(nil)
	}

	request, _ := http.NewRequest(method, path, reader)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, request)
	return recorder
}

func registerFin(t *testing.T, app *gin.Engine) fin.RegisterResponse {
	t.Helper()

	recorder := doRequest(app, "POST", "/fin/register", "", fin.RegisterRequest{
		RegistrationToken: registrationToken,
		DisplayName:       "test-fin",
		Capabilities:      []fin.Capability{{Type: "soarca-fin-test"}},
	})
	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response fin.RegisterResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("could not unmarshal register response: %v", err)
	}
	return response
}

func TestRegisterFin(t *testing.T) {
	app := newFinTestApp(t, registrationToken)

	response := registerFin(t, app)

	assert.NotEmpty(t, response.FinId)
	assert.NotEmpty(t, response.FinToken)
	assert.Equal(t, 5, response.PollIntervalSeconds)
	assert.Equal(t, 1, response.LongPollTimeoutSeconds)
	assert.Equal(t, 60, response.JobLeaseSeconds)
}

func TestRegisterFinRejectsWrongRegistrationToken(t *testing.T) {
	app := newFinTestApp(t, registrationToken)

	recorder := doRequest(app, "POST", "/fin/register", "", fin.RegisterRequest{
		RegistrationToken: "wrong-token",
		Capabilities:      []fin.Capability{{Type: "soarca-fin-test"}},
	})

	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestRegisterFinRejectsMissingCapabilities(t *testing.T) {
	app := newFinTestApp(t, registrationToken)

	recorder := doRequest(app, "POST", "/fin/register", "", fin.RegisterRequest{
		RegistrationToken: registrationToken,
	})

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestRegisterFinUnavailableWhenRegistrationTokenNotConfigured(t *testing.T) {
	app := newFinTestApp(t, "")

	recorder := doRequest(app, "POST", "/fin/register", "", fin.RegisterRequest{
		RegistrationToken: registrationToken,
		Capabilities:      []fin.Capability{{Type: "soarca-fin-test"}},
	})

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestPollRequiresAuthorizationHeader(t *testing.T) {
	app := newFinTestApp(t, registrationToken)

	recorder := doRequest(app, "POST", "/fin/poll", "", nil)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestPollRejectsUnknownToken(t *testing.T) {
	app := newFinTestApp(t, registrationToken)

	recorder := doRequest(app, "POST", "/fin/poll", "not-a-real-token", nil)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestPollReturnsNoContentWhenNoWorkAvailable(t *testing.T) {
	app := newFinTestApp(t, registrationToken)
	registered := registerFin(t, app)

	recorder := doRequest(app, "POST", "/fin/poll", registered.FinToken, fin.PollRequest{ConcurrencyAvailable: 1})

	assert.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestSubmitResultRejectsMalformedJobId(t *testing.T) {
	app := newFinTestApp(t, registrationToken)
	registered := registerFin(t, app)

	recorder := doRequest(app, "PUT", "/fin/jobs/not-a-uuid", registered.FinToken, fin.ResultRequest{
		JobResult: fin.JobResult{State: fin.JobStateSuccess},
	})

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestSubmitResultRejectsInvalidState(t *testing.T) {
	app := newFinTestApp(t, registrationToken)
	registered := registerFin(t, app)

	recorder := doRequest(app, "PUT", "/fin/jobs/6ba7b810-9dad-11d1-80b4-00c04fd430c0", registered.FinToken,
		fin.ResultRequest{JobResult: fin.JobResult{State: "not-a-state"}})

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestSubmitResultForUnknownJobIsNotFound(t *testing.T) {
	app := newFinTestApp(t, registrationToken)
	registered := registerFin(t, app)

	recorder := doRequest(app, "PUT", "/fin/jobs/6ba7b810-9dad-11d1-80b4-00c04fd430c0", registered.FinToken,
		fin.ResultRequest{JobResult: fin.JobResult{State: fin.JobStateSuccess}})

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestListFins(t *testing.T) {
	app := newFinTestApp(t, registrationToken)
	registered := registerFin(t, app)

	recorder := doRequest(app, "GET", "/fin/", "", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response fin.ListResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("could not unmarshal list response: %v", err)
	}

	assert.Equal(t, 1, len(response.Fins))
	assert.Equal(t, registered.FinId, response.Fins[0].FinId)
	assert.Equal(t, "test-fin", response.Fins[0].DisplayName)
	assert.Equal(t, false, response.Fins[0].Stale)
}

// The fin token hash is a credential; Record doubles as the admin wire type and
// only a json:"-" tag keeps the hash off the wire.
func TestListFinsDoesNotLeakTokenHash(t *testing.T) {
	app := newFinTestApp(t, registrationToken)
	registered := registerFin(t, app)

	recorder := doRequest(app, "GET", "/fin/", "", nil)
	body := recorder.Body.String()

	assert.Equal(t, false, strings.Contains(body, "fin_token_hash"))
	assert.Equal(t, false, strings.Contains(body, "FinTokenHash"))
	assert.Equal(t, false, strings.Contains(body, registered.FinToken))
}

func TestGetFinById(t *testing.T) {
	app := newFinTestApp(t, registrationToken)
	registered := registerFin(t, app)

	recorder := doRequest(app, "GET", "/fin/"+registered.FinId, "", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var record fin.Record
	if err := json.Unmarshal(recorder.Body.Bytes(), &record); err != nil {
		t.Fatalf("could not unmarshal fin record: %v", err)
	}
	assert.Equal(t, registered.FinId, record.FinId)
}

func TestGetUnknownFinIsNotFound(t *testing.T) {
	app := newFinTestApp(t, registrationToken)

	recorder := doRequest(app, "GET", "/fin/does-not-exist", "", nil)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestDeleteFin(t *testing.T) {
	app := newFinTestApp(t, registrationToken)
	registered := registerFin(t, app)

	recorder := doRequest(app, "DELETE", "/fin/"+registered.FinId, "", nil)
	assert.Equal(t, http.StatusNoContent, recorder.Code)

	recorder = doRequest(app, "GET", "/fin/"+registered.FinId, "", nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestUnregisterFin(t *testing.T) {
	app := newFinTestApp(t, registrationToken)
	registered := registerFin(t, app)

	recorder := doRequest(app, "DELETE", "/fin/", registered.FinToken, nil)
	assert.Equal(t, http.StatusNoContent, recorder.Code)

	recorder = doRequest(app, "POST", "/fin/poll", registered.FinToken, nil)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
