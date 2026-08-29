package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	finservice "soarca/internal/services/fin"
	"soarca/internal/storage/memory"
	"soarca/pkg/api/fin"
	"soarca/pkg/core/capability/fin/queue"
	finmodels "soarca/pkg/models/fin"
	"soarca/test/unittest/mocks/mock_guid"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

// requireAdminHeader is a stand-in for gauth's real JWT middleware
// (internal/controller/controller.go's intializeAuthenticationMiddleware) -
// it only exists to exercise gin's actual registration-order semantics
// without needing a real OIDC/JWKS setup in a unit test. It rejects any
// request that doesn't carry X-Admin-Test: yes.
func requireAdminHeader(g *gin.Context) {
	if g.GetHeader("X-Admin-Test") != "yes" {
		g.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	g.Next()
}

// This test pins the exact ordering hazard the Fin routes must avoid: gin's
// engine.Use() only affects routes registered *after* it is called - routes
// already registered keep whatever handler chain they had at registration
// time. FinPublicRoutes must therefore be registered before any global
// admin-auth middleware is installed (see FinPublic's doc comment and
// internal/controller/controller.go's ordering), while FinAdminRoutes is
// expected to sit behind it, like the rest of the admin API.
func TestFinPublicRoutesAreExemptFromAdminAuthButFinAdminRoutesAreNot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := gin.New()

	repository := memory.New().Fins()
	jobQueue := queue.New()
	t.Cleanup(jobQueue.Close)
	guidMock := new(mock_guid.Mock_Guid)
	guidMock.On("New").Return(uuid.New())

	finHandler := fin.NewFinHandler(
		finservice.NewRegistry(repository, finservice.RegistryConfig{
			RegistrationToken: "test-registration-token",
		}, guidMock),
		finservice.NewWorkService(repository, jobQueue, finservice.WorkServiceConfig{
			LongPollTimeoutSeconds: 1,
			JobLeaseSeconds:        60,
		}),
		fin.Config{
			RegistrationToken:      "test-registration-token",
			PollIntervalSeconds:    5,
			LongPollTimeoutSeconds: 1,
			JobLeaseSeconds:        60,
		},
	)

	// Simulates: routes.FinPublic(app, finHandler) called before
	// intializeAuthenticationMiddleware(app) in controller.go.
	FinPublicRoutes(app, finHandler)

	// Simulates: intializeAuthenticationMiddleware(app) installing the
	// global admin-auth middleware.
	app.Use(requireAdminHeader)

	// Simulates: routes.FinAdmin(app, finHandler) called afterwards,
	// alongside the other admin routes (routes.Api, routes.Manual, etc.).
	FinAdminRoutes(app, finHandler)

	registerRequest := httptest.NewRequest(http.MethodPost, "/fin/register", jsonBody(t, finmodels.RegisterRequest{
		RegistrationToken: "test-registration-token",
		Capabilities:      []finmodels.Capability{{Type: "pong"}},
	}))
	registerRequest.Header.Set("Content-Type", "application/json")
	registerRecorder := httptest.NewRecorder()
	app.ServeHTTP(registerRecorder, registerRequest)
	assert.Equal(t, registerRecorder.Code, http.StatusCreated)

	listRequest := httptest.NewRequest(http.MethodGet, "/fin/", nil)
	listRecorder := httptest.NewRecorder()
	app.ServeHTTP(listRecorder, listRequest)
	if listRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected GET /fin/ to require admin auth, got %d", listRecorder.Code)
	}

	listRequest = httptest.NewRequest(http.MethodGet, "/fin/", nil)
	listRequest.Header.Set("X-Admin-Test", "yes")
	listRecorder = httptest.NewRecorder()
	app.ServeHTTP(listRecorder, listRequest)
	assert.Equal(t, listRecorder.Code, http.StatusOK)
}

func jsonBody(t *testing.T, v any) io.Reader {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(data)
}
