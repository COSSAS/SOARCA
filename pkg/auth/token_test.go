package auth_test

import (
	"context"
	"testing"
	"time"

	"soarca/pkg/auth"
	"soarca/test/unittest/fakes/fake_identity_provider"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-playground/assert/v2"
)

const audience = "soarca-api"

func newValidator(t *testing.T) (*auth.TokenValidator, *fake_identity_provider.IdentityProvider) {
	provider := fake_identity_provider.New(t, false)
	provider.ServeKeys("/keys")
	keys := oidc.NewRemoteKeySet(context.Background(), provider.Url("/keys"))
	return auth.NewTokenValidator(provider.Url("/"), keys, []string{audience}, nil), provider
}

func TestValidateAcceptsAValidToken(t *testing.T) {
	validator, provider := newValidator(t)

	token, err := validator.Validate(context.Background(), provider.Sign(t, map[string]any{
		"iss": provider.Url("/"), "aud": audience, "sub": "caller", "name": "Caller",
	}))

	assert.Equal(t, err, nil)
	assert.Equal(t, token.Subject, "caller")
	var claims struct {
		Name string `json:"name"`
	}
	assert.Equal(t, token.Claims(&claims), nil)
	assert.Equal(t, claims.Name, "Caller")
}

func TestValidateRejects(t *testing.T) {
	validator, provider := newValidator(t)
	valid := map[string]any{"iss": provider.Url("/"), "aud": audience, "sub": "caller"}
	with := func(name string, value any) map[string]any {
		claims := map[string]any{}
		for key, existing := range valid {
			claims[key] = existing
		}
		claims[name] = value
		return claims
	}

	for name, token := range map[string]string{
		"wrong audience":  provider.Sign(t, with("aud", "graph")),
		"wrong issuer":    provider.Sign(t, with("iss", "https://evil.example/")),
		"expired":         provider.Sign(t, with("exp", time.Now().Add(-time.Hour).Unix())),
		"unpublished key": provider.SignWithUnknownKey(t, valid),
	} {
		_, err := validator.Validate(context.Background(), token)
		if err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestValidateAllowsClockLeeway(t *testing.T) {
	validator, provider := newValidator(t)

	_, err := validator.Validate(context.Background(), provider.Sign(t, map[string]any{
		"iss": provider.Url("/"), "aud": audience, "exp": time.Now().Add(-auth.ClockLeeway / 2).Unix(),
	}))

	assert.Equal(t, err, nil)
}
