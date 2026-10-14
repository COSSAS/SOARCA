package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"soarca/pkg/auth"
	"soarca/pkg/auth/roles"
	"soarca/test/unittest/mocks/mock_auth"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/mock"
)

var executor = &auth.Principal{
	Subjects:    []string{"entra:tenant:app:terminal"},
	Permissions: map[roles.Permission]bool{roles.PlaybookRead: true, roles.PlaybookExecute: true},
}

var reader = &auth.Principal{
	Subjects:    []string{"authentik:user:reader"},
	Permissions: map[roles.Permission]bool{roles.PlaybookRead: true},
}

func newRouter(authenticator auth.IAuthenticator) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes := router.Group("/", Middleware(authenticator))
	routes.GET("/playbook/", RequireAll(roles.PlaybookRead), func(ginContext *gin.Context) {
		ginContext.String(http.StatusOK, PrincipalFromContext(ginContext).Self())
	})
	routes.POST("/trigger/playbook", RequireAll(roles.PlaybookExecute), func(ginContext *gin.Context) {
		ginContext.Status(http.StatusOK)
	})
	routes.GET("/either", RequireAny(roles.PlaybookWrite, roles.PlaybookExecute), func(ginContext *gin.Context) {
		ginContext.Status(http.StatusOK)
	})
	return router
}

func serve(router *gin.Engine, method string, path string, authorization string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func authenticatorFor(credentials map[string]*auth.Principal) *mock_auth.MockAuthenticator {
	authenticator := &mock_auth.MockAuthenticator{}
	for credential, principal := range credentials {
		authenticator.On("Authenticate", mock.Anything, credential).Return(principal, nil)
	}
	authenticator.On("Authenticate", mock.Anything, mock.Anything).Return(nil, auth.ErrUnauthenticated)
	return authenticator
}

func TestPermissionsDecideAccess(t *testing.T) {
	router := newRouter(authenticatorFor(map[string]*auth.Principal{"executor": executor, "reader": reader}))

	assert.Equal(t, serve(router, http.MethodPost, "/trigger/playbook", "Bearer executor").Code, http.StatusOK)
	assert.Equal(t, serve(router, http.MethodGet, "/playbook/", "Bearer executor").Code, http.StatusOK)
	assert.Equal(t, serve(router, http.MethodPost, "/trigger/playbook", "Bearer reader").Code, http.StatusForbidden)
	assert.Equal(t, serve(router, http.MethodGet, "/either", "Bearer executor").Code, http.StatusOK)
	assert.Equal(t, serve(router, http.MethodGet, "/either", "Bearer reader").Code, http.StatusForbidden)
}

func TestHandlersSeeThePrincipal(t *testing.T) {
	router := newRouter(authenticatorFor(map[string]*auth.Principal{"reader": reader}))

	assert.Equal(t, serve(router, http.MethodGet, "/playbook/", "Bearer reader").Body.String(), "authentik:user:reader")
}

func TestFailedAuthenticationIs401(t *testing.T) {
	authenticator := &mock_auth.MockAuthenticator{}
	authenticator.On("Authenticate", mock.Anything, "bad").Return(nil, errors.New("bad signature"))
	authenticator.On("Authenticate", mock.Anything, "").Return(nil, auth.ErrUnauthenticated)
	router := newRouter(authenticator)

	assert.Equal(t, serve(router, http.MethodGet, "/playbook/", "Bearer bad").Code, http.StatusUnauthorized)
	assert.Equal(t, serve(router, http.MethodGet, "/playbook/", "").Code, http.StatusUnauthorized)
	assert.Equal(t, serve(router, http.MethodGet, "/playbook/", "Basic dXNlcjpwYXNz").Code, http.StatusUnauthorized)
}

func TestErrorUsesTheSoarcaErrorFormat(t *testing.T) {
	router := newRouter(authenticatorFor(nil))

	recorder := serve(router, http.MethodGet, "/playbook/", "")

	assert.Equal(t, recorder.Header().Get("WWW-Authenticate"), `Bearer realm="soarca"`)
	var body map[string]string
	assert.Equal(t, json.Unmarshal(recorder.Body.Bytes(), &body), nil)
	assert.Equal(t, body["status"], "401")
	assert.Equal(t, body["message"], "unauthenticated")
	assert.Equal(t, body["original-call"], "GET /playbook/")
}

func TestRequireWithoutMiddlewareIs401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/playbook/", RequireAll(roles.PlaybookRead), func(ginContext *gin.Context) {
		ginContext.Status(http.StatusOK)
	})

	assert.Equal(t, serve(router, http.MethodGet, "/playbook/", "").Code, http.StatusUnauthorized)
}

func TestBearerCredential(t *testing.T) {
	assert.Equal(t, bearerCredential("Bearer abc"), "abc")
	assert.Equal(t, bearerCredential("bearer  abc "), "abc")
	assert.Equal(t, bearerCredential("Basic abc"), "")
	assert.Equal(t, bearerCredential("Bearer"), "")
	assert.Equal(t, bearerCredential(""), "")
}
