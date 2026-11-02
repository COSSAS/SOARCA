package auth

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"soarca/pkg/auth"
	"soarca/pkg/auth/roles"
	"soarca/pkg/models/api"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func getMetadata(t *testing.T, handler *AuthHandler, request *http.Request) (map[string]any, api.ProtectedResourceMetadata) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET(ProtectedResourcePath, handler.GetProtectedResourceMetadata)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	assert.Equal(t, recorder.Code, http.StatusOK)

	var raw map[string]any
	assert.Equal(t, json.Unmarshal(recorder.Body.Bytes(), &raw), nil)
	var metadata api.ProtectedResourceMetadata
	assert.Equal(t, json.Unmarshal(recorder.Body.Bytes(), &metadata), nil)
	return raw, metadata
}

func TestMetadataListsIssuersAndScopes(t *testing.T) {
	handler := NewAuthHandler(
		[]string{"https://authentik.example.com/application/o/soarca/", "https://login.microsoftonline.com/tenant/v2.0"},
		[]string{"api://soarca/access_as_user"},
		"",
	)
	request := httptest.NewRequest(http.MethodGet, "https://soarca.example.com"+ProtectedResourcePath, nil)
	request.TLS = &tls.ConnectionState{}

	_, metadata := getMetadata(t, handler, request)
	assert.Equal(t, metadata, api.ProtectedResourceMetadata{
		Resource:               "https://soarca.example.com",
		ResourceName:           "SOARCA",
		AuthorizationServers:   []string{"https://authentik.example.com/application/o/soarca/", "https://login.microsoftonline.com/tenant/v2.0"},
		ScopesSupported:        []string{"api://soarca/access_as_user"},
		BearerMethodsSupported: []string{"header"},
	})
}

func TestMetadataUsesTheConfiguredResource(t *testing.T) {
	handler := NewAuthHandler([]string{"https://issuer.example/"}, nil, "https://soarca.example.com")
	request := httptest.NewRequest(http.MethodGet, "http://10.0.0.5:8080"+ProtectedResourcePath, nil)

	_, metadata := getMetadata(t, handler, request)
	assert.Equal(t, metadata.Resource, "https://soarca.example.com")
}

func TestMetadataWithoutAuthenticationListsNoIssuers(t *testing.T) {
	handler := NewAuthHandler(nil, nil, "")
	raw, _ := getMetadata(t, handler, httptest.NewRequest(http.MethodGet, ProtectedResourcePath, nil))

	// An empty list, not null, tells clients they can skip sign-in.
	assert.Equal(t, raw["authorization_servers"], []any{})
	_, hasScopes := raw["scopes_supported"]
	assert.Equal(t, hasScopes, false)
}

func TestUnauthenticatedResponseUsesTheConfiguredResource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Protect(authenticatorFor(nil), testPolicy, "https://soarca.example.com"))
	router.GET("/playbook/:id", func(ginContext *gin.Context) { ginContext.Status(http.StatusOK) })

	recorder := serve(router, http.MethodGet, "/playbook/1", "")
	assert.Equal(t, recorder.Code, http.StatusUnauthorized)
	assert.Equal(t, recorder.Header().Get("WWW-Authenticate"),
		`Bearer realm="soarca", resource_metadata="https://soarca.example.com/.well-known/oauth-protected-resource"`)
}

func getPermissions(t *testing.T, authenticator auth.IAuthenticator, authorization string) (int, api.PrincipalPermissions) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Protect(authenticator, testPolicy, ""))
	router.GET("/principal/permissions", NewAuthHandler(nil, nil, "").GetPrincipalPermissions)
	recorder := serve(router, http.MethodGet, "/principal/permissions", authorization)

	var body api.PrincipalPermissions
	if recorder.Code == http.StatusOK {
		assert.Equal(t, recorder.Header().Get("Cache-Control"), "no-store")
		assert.Equal(t, json.Unmarshal(recorder.Body.Bytes(), &body), nil)
	}
	return recorder.Code, body
}

func TestPrincipalPermissionsListsWhatTheCallerHolds(t *testing.T) {
	executorWithProvider := *executor
	executorWithProvider.Provider = "entra"
	authenticator := authenticatorFor(map[string]*auth.Principal{
		"executor": &executorWithProvider,
		"admin":    {Provider: "authentik", Subjects: []string{"authentik:user:admin"}, Permissions: map[roles.Permission]bool{roles.Admin: true}},
		"nobody":   {Provider: "apikey", Subjects: []string{"apikey:app:nobody"}},
	})

	code, body := getPermissions(t, authenticator, "Bearer executor")
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, body, api.PrincipalPermissions{
		Subject:     "entra:tenant:app:terminal",
		Provider:    "entra",
		Permissions: []string{"playbook:read", "playbook:execute"},
	})

	// soarca:admin is expanded, so clients don't need to know what it implies.
	_, body = getPermissions(t, authenticator, "Bearer admin")
	assert.Equal(t, body.Permissions, []string{"playbook:read", "playbook:write", "playbook:execute", "soarca:admin"})

	// Without any permission the answer is a 403, which tells a client the
	// caller signed in but has no access to SOARCA.
	code, _ = getPermissions(t, authenticator, "Bearer nobody")
	assert.Equal(t, code, http.StatusForbidden)

	code, _ = getPermissions(t, authenticator, "")
	assert.Equal(t, code, http.StatusUnauthorized)
}
