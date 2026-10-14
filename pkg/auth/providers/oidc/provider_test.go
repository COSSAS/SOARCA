package oidc

import (
	"context"
	"testing"

	"soarca/pkg/auth"
	"soarca/test/unittest/fakes/fake_identity_provider"

	"github.com/go-playground/assert/v2"
)

const issuerPath = "/application/o/soarca/"

func fakeAuthentik(t *testing.T, useTls bool) *fake_identity_provider.IdentityProvider {
	server := fake_identity_provider.New(t, useTls)
	server.ServeDiscovery(issuerPath)
	return server
}

func credential(t *testing.T, server *fake_identity_provider.IdentityProvider, claims map[string]any) auth.Credential {
	return auth.Credential{Issuer: server.Url(issuerPath), Raw: server.Sign(t, claims)}
}

func TestNameAndHandles(t *testing.T) {
	server := fakeAuthentik(t, false)
	provider, err := New(Config{Name: "authentik", IssuerUrl: server.Url(issuerPath), Audiences: []string{"soarca"}})
	assert.Equal(t, err, nil)

	assert.Equal(t, provider.Name(), "authentik")
	assert.Equal(t, provider.Handles(), auth.Match{Issuers: []string{server.Url(issuerPath)}})
}

func TestAuthenticateThroughDiscovery(t *testing.T) {
	server := fakeAuthentik(t, false)
	provider, err := New(Config{Name: "authentik", IssuerUrl: server.Url(issuerPath), Audiences: []string{"soarca"}})
	assert.Equal(t, err, nil)

	principal, err := provider.Authenticate(context.Background(), credential(t, server, map[string]any{
		"iss": server.Url(issuerPath), "aud": "soarca", "sub": "user-1", "groups": []string{"soarca_read"},
	}))

	assert.Equal(t, err, nil)
	assert.Equal(t, principal.Self(), "authentik:user:user-1")
}

func TestAuthenticateRejectsOtherAudience(t *testing.T) {
	server := fakeAuthentik(t, false)
	provider, err := New(Config{IssuerUrl: server.Url(issuerPath), Audiences: []string{"soarca"}})
	assert.Equal(t, err, nil)

	_, err = provider.Authenticate(context.Background(), credential(t, server, map[string]any{
		"iss": server.Url(issuerPath), "aud": "another-app", "sub": "user-1",
	}))

	assert.NotEqual(t, err, nil)
}

func TestSkipTlsVerify(t *testing.T) {
	server := fakeAuthentik(t, true) // self-signed certificate
	token := credential(t, server, map[string]any{"iss": server.Url(issuerPath), "aud": "soarca", "sub": "user-1"})

	strict, err := New(Config{IssuerUrl: server.Url(issuerPath), Audiences: []string{"soarca"}})
	assert.Equal(t, err, nil)
	_, err = strict.Authenticate(context.Background(), token)
	assert.NotEqual(t, err, nil)

	lenient, err := New(Config{IssuerUrl: server.Url(issuerPath), Audiences: []string{"soarca"}, SkipTlsVerify: true})
	assert.Equal(t, err, nil)
	_, err = lenient.Authenticate(context.Background(), token)
	assert.Equal(t, err, nil)
}

func TestDiscoveryIsRetriedAfterFailure(t *testing.T) {
	server := fake_identity_provider.New(t, false)
	provider, err := New(Config{IssuerUrl: server.Url(issuerPath), Audiences: []string{"soarca"}})
	assert.Equal(t, err, nil)
	token := credential(t, server, map[string]any{"iss": server.Url(issuerPath), "aud": "soarca", "sub": "user-1"})

	_, err = provider.Authenticate(context.Background(), token)
	assert.NotEqual(t, err, nil)

	server.ServeDiscovery(issuerPath)
	_, err = provider.Authenticate(context.Background(), token)
	assert.Equal(t, err, nil)
}

func TestNewRejectsIncompleteConfig(t *testing.T) {
	_, err := New(Config{Audiences: []string{"soarca"}})
	assert.NotEqual(t, err, nil)

	_, err = New(Config{IssuerUrl: "https://idp/"})
	assert.NotEqual(t, err, nil)
}
