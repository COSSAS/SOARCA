package auth_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"soarca/pkg/auth"
	"soarca/test/unittest/mocks/mock_auth"

	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/mock"
)

// unsignedJwt builds a token with only an issuer; routing never checks signatures.
func unsignedJwt(issuer string) string {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"` + issuer + `"}`))
	return "eyJhbGciOiJub25lIn0." + payload + ".sig"
}

func provider(name string, match auth.Match) *mock_auth.MockIdentityProvider {
	identityProvider := &mock_auth.MockIdentityProvider{}
	identityProvider.On("Name").Return(name)
	identityProvider.On("Handles").Return(match)
	return identityProvider
}

func TestRoutesJwtToTheProviderForItsIssuer(t *testing.T) {
	entra := provider("entra", auth.Match{Issuers: []string{"https://login.example/tenant/v2.0"}})
	authentik := provider("authentik", auth.Match{Issuers: []string{"https://authentik.example/application/o/soarca/"}})
	token := unsignedJwt("https://authentik.example/application/o/soarca/")
	authentik.On("Authenticate", mock.Anything, auth.Credential{Raw: token, Issuer: "https://authentik.example/application/o/soarca/"}).
		Return(&auth.Principal{Subject: "user-1"}, nil)

	authenticator, err := auth.New([]auth.IIdentityProvider{entra, authentik})
	assert.Equal(t, err, nil)
	principal, err := authenticator.Authenticate(context.Background(), token)

	assert.Equal(t, err, nil)
	assert.Equal(t, principal.Subject, "user-1")
	assert.Equal(t, principal.Provider, "authentik")
	entra.AssertNotCalled(t, "Authenticate", mock.Anything, mock.Anything)
}

func TestRoutesPrefixedCredentialsToTheirProvider(t *testing.T) {
	jwt := provider("entra", auth.Match{Issuers: []string{"https://login.example/tenant/v2.0"}})
	apiKeys := provider("apikey", auth.Match{Prefix: "soarca_"})
	apiKeys.On("Authenticate", mock.Anything, auth.Credential{Raw: "soarca_abc"}).Return(&auth.Principal{Subject: "webhook"}, nil)

	authenticator, err := auth.New([]auth.IIdentityProvider{jwt, apiKeys})
	assert.Equal(t, err, nil)
	principal, err := authenticator.Authenticate(context.Background(), "soarca_abc")

	assert.Equal(t, err, nil)
	assert.Equal(t, principal.Provider, "apikey")
}

func TestProviderErrorsAreUnauthenticated(t *testing.T) {
	failing := provider("authentik", auth.Match{Issuers: []string{"https://issuer.example/"}})
	token := unsignedJwt("https://issuer.example/")
	failing.On("Authenticate", mock.Anything, mock.Anything).Return(nil, errors.New("bad signature"))

	authenticator, err := auth.New([]auth.IIdentityProvider{failing})
	assert.Equal(t, err, nil)
	_, err = authenticator.Authenticate(context.Background(), token)

	assert.Equal(t, errors.Is(err, auth.ErrUnauthenticated), true)
}

func TestRejectsCredentialsNoProviderHandles(t *testing.T) {
	authenticator, err := auth.New([]auth.IIdentityProvider{provider("authentik", auth.Match{Issuers: []string{"https://issuer.example/"}})})
	assert.Equal(t, err, nil)

	for name, credential := range map[string]string{
		"empty":          "",
		"not a jwt":      "abc",
		"bad payload":    "a.!!!.c",
		"no issuer":      "a." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"x"}`)) + ".c",
		"unknown issuer": unsignedJwt("https://evil.example/"),
	} {
		_, err := authenticator.Authenticate(context.Background(), credential)
		if !errors.Is(err, auth.ErrUnauthenticated) {
			t.Errorf("%s: got %v, want ErrUnauthenticated", name, err)
		}
	}
}

func TestNewRejectsConflictingProviders(t *testing.T) {
	issuer := auth.Match{Issuers: []string{"https://a/"}}

	for name, providers := range map[string][]auth.IIdentityProvider{
		"none":               nil,
		"empty name":         {provider("", issuer)},
		"duplicate name":     {provider("a", issuer), provider("a", auth.Match{Prefix: "x_"})},
		"handles nothing":    {provider("a", auth.Match{})},
		"duplicate issuer":   {provider("a", issuer), provider("b", issuer)},
		"overlapping prefix": {provider("a", auth.Match{Prefix: "soarca_"}), provider("b", auth.Match{Prefix: "soarca_x"})},
	} {
		if _, err := auth.New(providers); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestPeekIssuer(t *testing.T) {
	issuer, err := auth.PeekIssuer(unsignedJwt("https://issuer.example/"))

	assert.Equal(t, err, nil)
	assert.Equal(t, issuer, "https://issuer.example/")
}
