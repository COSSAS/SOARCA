package controller

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"soarca/internal/config"
	"soarca/pkg/auth"
	"soarca/pkg/auth/providers/apikey"
	"soarca/pkg/auth/roles"

	"github.com/go-playground/assert/v2"
)

func writeKeys(t *testing.T, content string) string {
	path := filepath.Join(t.TempDir(), "api-keys.json")
	assert.Equal(t, os.WriteFile(path, []byte(content), 0o600), nil)
	return path
}

func TestDisabledAuthIsAnonymous(t *testing.T) {
	authenticator, err := newAuthenticator(config.Auth{Enabled: false, ApiKeysFile: "/does/not/exist"})
	assert.Equal(t, err, nil)

	principal, err := authenticator.Authenticate(context.Background(), "")
	assert.Equal(t, err, nil)
	assert.Equal(t, principal.Provider, auth.AnonymousName)
	assert.Equal(t, principal.Has(roles.Admin), true)
}

func TestApiKeysFromFile(t *testing.T) {
	plaintext, hash, err := apikey.Generate()
	assert.Equal(t, err, nil)
	path := writeKeys(t, `[{"name":"webhook","sha256":"`+hash+`","permissions":["playbook:execute"],"expires":"2999-01-01T00:00:00Z"}]`)

	authenticator, err := newAuthenticator(config.Auth{Enabled: true, ApiKeysFile: path})
	assert.Equal(t, err, nil)

	principal, err := authenticator.Authenticate(context.Background(), plaintext)
	assert.Equal(t, err, nil)
	assert.Equal(t, principal.Provider, apikey.Name)
	assert.Equal(t, principal.Has(roles.PlaybookExecute), true)
	assert.Equal(t, principal.Has(roles.PlaybookWrite), false)
}

func TestInvalidApiKeysFileStopsStartup(t *testing.T) {
	_, err := newAuthenticator(config.Auth{Enabled: true, ApiKeysFile: "/does/not/exist"})
	assert.NotEqual(t, err, nil)

	_, err = newAuthenticator(config.Auth{Enabled: true, ApiKeysFile: writeKeys(t, `[{"name":"no-expiry","sha256":"x"}]`)})
	assert.NotEqual(t, err, nil)
}

func TestEveryProviderTogether(t *testing.T) {
	_, hash, _ := apikey.Generate()
	authConfig := config.Auth{
		Enabled:    true,
		Issuer:     "https://authentik.example.com/application/o/soarca/",
		ClientId:   "soarca",
		AdminGroup: "soarca_admin",
		Entra: config.Entra{
			TenantIds:         []string{"11111111-1111-1111-1111-111111111111"},
			ApiClientId:       "aaaaaaaa-0000-0000-0000-000000000001",
			RequiredUserScope: "access_as_user",
			Authority:         "https://login.microsoftonline.com",
		},
		ApiKeysFile: writeKeys(t, `[{"name":"webhook","sha256":"`+hash+`","permissions":["playbook:read"],"expires":"2999-01-01T00:00:00Z"}]`),
	}

	authenticator, err := newAuthenticator(authConfig)
	assert.Equal(t, err, nil)

	// Neither identity provider is contacted at startup, and a credential
	// no provider handles is rejected without reaching out either.
	_, err = authenticator.Authenticate(context.Background(), "")
	assert.NotEqual(t, err, nil)
}

func TestAdminApiKeyReachesEveryPermission(t *testing.T) {
	plaintext, hash, _ := apikey.Generate()
	path := writeKeys(t, `[{"name":"admin","sha256":"`+hash+`","permissions":["soarca:admin"],"expires":"2999-01-01T00:00:00Z"}]`)

	authenticator, err := newAuthenticator(config.Auth{Enabled: true, ApiKeysFile: path})
	assert.Equal(t, err, nil)

	principal, err := authenticator.Authenticate(context.Background(), plaintext)
	assert.Equal(t, err, nil)
	for _, permission := range roles.All() {
		assert.Equal(t, principal.Has(permission), true)
	}
}
