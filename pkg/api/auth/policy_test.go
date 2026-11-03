package auth

import (
	"net/http"
	"testing"

	"soarca/pkg/auth"
	"soarca/pkg/auth/roles"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

var admin = &auth.Principal{
	Subjects:    []string{"authentik:user:admin"},
	Permissions: map[roles.Permission]bool{roles.Admin: true, roles.PlaybookRead: true},
}

var testPolicy = Policy{
	RouteKey(http.MethodGet, "/status/ping"):       {Public: true},
	RouteKey(http.MethodGet, "/playbook/:id"):      {Permission: roles.PlaybookRead},
	RouteKey(http.MethodPost, "/trigger/playbook"): {Permission: roles.PlaybookExecute},
}

func newProtectedRouter(authenticator auth.IAuthenticator) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Protect(authenticator, testPolicy))
	ok := func(ginContext *gin.Context) {
		ginContext.String(http.StatusOK, PrincipalFromContext(ginContext).Self())
	}
	router.GET("/status/ping", func(ginContext *gin.Context) { ginContext.String(http.StatusOK, "pong") })
	router.GET("/playbook/:id", ok)
	router.POST("/trigger/playbook", ok)
	router.GET("/unlisted", ok)
	return router
}

func TestPublicRouteNeedsNoCredential(t *testing.T) {
	router := newProtectedRouter(authenticatorFor(nil))

	recorder := serve(router, http.MethodGet, "/status/ping", "")
	assert.Equal(t, recorder.Code, http.StatusOK)
	assert.Equal(t, recorder.Body.String(), "pong")
}

func TestProtectedRouteNeedsCredentialAndPermission(t *testing.T) {
	router := newProtectedRouter(authenticatorFor(map[string]*auth.Principal{"executor": executor, "reader": reader}))

	assert.Equal(t, serve(router, http.MethodGet, "/playbook/abc", "").Code, http.StatusUnauthorized)
	assert.Equal(t, serve(router, http.MethodGet, "/playbook/abc", "Bearer forged").Code, http.StatusUnauthorized)
	assert.Equal(t, serve(router, http.MethodPost, "/trigger/playbook", "Bearer reader").Code, http.StatusForbidden)

	recorder := serve(router, http.MethodGet, "/playbook/abc", "Bearer reader")
	assert.Equal(t, recorder.Code, http.StatusOK)
	assert.Equal(t, recorder.Body.String(), "authentik:user:reader")
	assert.Equal(t, serve(router, http.MethodPost, "/trigger/playbook", "Bearer executor").Code, http.StatusOK)
}

func TestRouteWithoutRuleNeedsAdmin(t *testing.T) {
	router := newProtectedRouter(authenticatorFor(map[string]*auth.Principal{"executor": executor, "admin": admin}))

	assert.Equal(t, serve(router, http.MethodGet, "/unlisted", "Bearer executor").Code, http.StatusForbidden)
	assert.Equal(t, serve(router, http.MethodGet, "/unlisted", "Bearer admin").Code, http.StatusOK)
}

func TestUnknownPathNeedsCredentialBefore404(t *testing.T) {
	router := newProtectedRouter(authenticatorFor(map[string]*auth.Principal{"reader": reader}))

	assert.Equal(t, serve(router, http.MethodGet, "/does-not-exist", "").Code, http.StatusUnauthorized)
	assert.Equal(t, serve(router, http.MethodGet, "/does-not-exist", "Bearer reader").Code, http.StatusNotFound)
}

func TestRuleIsPerMethod(t *testing.T) {
	router := newProtectedRouter(authenticatorFor(nil))
	router.POST("/status/ping", func(ginContext *gin.Context) { ginContext.Status(http.StatusOK) })

	assert.Equal(t, serve(router, http.MethodPost, "/status/ping", "").Code, http.StatusUnauthorized)
}
