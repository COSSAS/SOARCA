package entra

import (
	"context"
	"testing"

	"soarca/pkg/auth"
	"soarca/test/unittest/fakes/fake_identity_provider"

	"github.com/go-playground/assert/v2"
)

const otherTenant = "22222222-2222-2222-2222-222222222222"

// fakeEntra serves key sets for tenant and otherTenant at Entra's paths.
func fakeEntra(t *testing.T) *fake_identity_provider.IdentityProvider {
	server := fake_identity_provider.New(t, false)
	server.ServeKeys("/" + tenant + "/discovery/v2.0/keys")
	server.ServeKeys("/" + otherTenant + "/discovery/v2.0/keys")
	return server
}

func newTestProvider(t *testing.T, server *fake_identity_provider.IdentityProvider, acceptV1 bool) *Provider {
	provider, err := New(Config{
		TenantIds:      []string{tenant},
		ApiClientId:    "api-client-id",
		ApiAppIdUri:    "api://soarca",
		AcceptV1Tokens: acceptV1,
		Authority:      server.Url(""),
	})
	assert.Equal(t, err, nil)
	return provider
}

func TestNameAndHandles(t *testing.T) {
	server := fakeEntra(t)

	assert.Equal(t, newTestProvider(t, server, false).Name(), "entra")
	assert.Equal(t, newTestProvider(t, server, false).Handles(), auth.Match{Issuers: []string{server.Url("/" + tenant + "/v2.0")}})
	assert.Equal(t, newTestProvider(t, server, true).Handles(), auth.Match{Issuers: []string{
		server.Url("/" + tenant + "/v2.0"),
		"https://sts.windows.net/" + tenant + "/",
	}})
}

func TestAuthenticateSignedTokens(t *testing.T) {
	server := fakeEntra(t)
	provider := newTestProvider(t, server, true)
	appToken := func(issuer string, audience string) auth.Credential {
		return auth.Credential{Issuer: issuer, Raw: server.Sign(t, map[string]any{
			"iss": issuer, "aud": audience, "tid": tenant, "azp": "caller", "idtyp": "app", "roles": []string{"Playbook.Read"},
		})}
	}
	v2Issuer := server.Url("/" + tenant + "/v2.0")
	otherIssuer := server.Url("/" + otherTenant + "/v2.0")

	accepted := map[string]auth.Credential{
		"v2 token":               appToken(v2Issuer, "api-client-id"),
		"app id uri as audience": appToken(v2Issuer, "api://soarca"),
		"v1 token when accepted": appToken("https://sts.windows.net/"+tenant+"/", "api://soarca"),
	}
	for name, credential := range accepted {
		principal, err := provider.Authenticate(context.Background(), credential)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		assert.Equal(t, principal.Kind, auth.KindApp)
	}

	rejected := map[string]auth.Credential{
		"graph token":                 appToken(v2Issuer, "https://graph.microsoft.com"),
		"tenant not allowed":          appToken(otherIssuer, "api-client-id"),
		"issuer differs from routing": {Issuer: v2Issuer, Raw: appToken(otherIssuer, "api-client-id").Raw},
		"unpublished key": {Issuer: v2Issuer, Raw: server.SignWithUnknownKey(t, map[string]any{
			"iss": v2Issuer, "aud": "api-client-id", "tid": tenant, "roles": []string{"SOARCA.Admin"},
		})},
	}
	for name, credential := range rejected {
		if _, err := provider.Authenticate(context.Background(), credential); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestNewRejectsIncompleteConfig(t *testing.T) {
	_, err := New(Config{ApiClientId: "api"})
	assert.NotEqual(t, err, nil)

	_, err = New(Config{TenantIds: []string{tenant}})
	assert.NotEqual(t, err, nil)
}
