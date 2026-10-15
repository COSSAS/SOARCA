package auth

import (
	"context"

	"soarca/pkg/auth/roles"
)

// AnonymousName labels the principal Anonymous returns.
const AnonymousName = "anonymous"

// Anonymous accepts every request, with or without a credential, as a caller
// that holds every permission. SOARCA uses it when authentication is
// disabled, so handlers can always read the caller from the request.
type Anonymous struct{}

func (anonymous Anonymous) Authenticate(ctx context.Context, credential string) (*Principal, error) {
	permissions := map[roles.Permission]bool{}
	for _, permission := range roles.All() {
		permissions[permission] = true
	}
	return &Principal{
		Provider:    AnonymousName,
		Subject:     AnonymousName,
		Permissions: permissions,
		Subjects:    []string{AnonymousName},
	}, nil
}
