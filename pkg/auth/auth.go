// Package auth authenticates callers of the SOARCA API.
//
// Identity providers (Entra ID, generic OIDC, API keys; see the providers
// sub-packages) declare which credentials they handle. The Authenticator
// routes each credential to the one provider that handles it and returns a
// Principal. Permissions are defined in the roles package.
package auth

import (
	"context"
	"errors"
)

// ErrUnauthenticated wraps every authentication failure. The wrapped error
// explains why, for logs; it must not be sent to the client.
var ErrUnauthenticated = errors.New("unauthenticated")

// IAuthenticator turns a raw credential into the calling Principal.
type IAuthenticator interface {
	Authenticate(ctx context.Context, credential string) (*Principal, error)
}

// IIdentityProvider verifies the credentials it declares in Handles.
type IIdentityProvider interface {
	// Name labels the provider; the Authenticator puts it in
	// Principal.Provider.
	Name() string
	// Handles declares which credentials this provider verifies. It is read
	// once, when the Authenticator is built.
	Handles() Match
	// Authenticate verifies credential and maps it to a Principal.
	Authenticate(ctx context.Context, credential Credential) (*Principal, error)
}

// Match describes the credentials a provider handles: JWTs from any of
// Issuers, and opaque credentials that start with Prefix.
type Match struct {
	Issuers []string
	Prefix  string
}

// Credential is a raw credential as routed by the Authenticator. Issuer is
// the unverified "iss" claim of a JWT and empty for prefixed credentials;
// providers must still verify it.
type Credential struct {
	Raw    string
	Issuer string
}
