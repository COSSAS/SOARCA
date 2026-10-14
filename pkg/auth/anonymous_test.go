package auth_test

import (
	"context"
	"testing"

	"soarca/pkg/auth"
	"soarca/pkg/auth/roles"

	"github.com/go-playground/assert/v2"
)

func TestAnonymousHoldsEveryPermission(t *testing.T) {
	principal, err := auth.Anonymous{}.Authenticate(context.Background(), "")
	assert.Equal(t, err, nil)
	assert.Equal(t, principal.Provider, auth.AnonymousName)
	assert.Equal(t, principal.Self(), auth.AnonymousName)
	for _, permission := range roles.All() {
		assert.Equal(t, principal.Has(permission), true)
	}
}

func TestAnonymousReturnsAFreshPrincipal(t *testing.T) {
	first, _ := auth.Anonymous{}.Authenticate(context.Background(), "")
	first.Permissions[roles.Admin] = false

	second, _ := auth.Anonymous{}.Authenticate(context.Background(), "")
	assert.Equal(t, second.Has(roles.Admin), true)
}
