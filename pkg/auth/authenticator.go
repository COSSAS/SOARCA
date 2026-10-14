package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Authenticator routes each credential to the provider that handles it.
type Authenticator struct {
	byIssuer map[string]IIdentityProvider
	byPrefix map[string]IIdentityProvider
}

// New builds an Authenticator. It fails when providers conflict: the same
// name or issuer twice, or prefixes where one starts with the other.
func New(providers []IIdentityProvider) (*Authenticator, error) {
	authenticator := &Authenticator{
		byIssuer: map[string]IIdentityProvider{},
		byPrefix: map[string]IIdentityProvider{},
	}
	names := map[string]bool{}
	for _, provider := range providers {
		name := provider.Name()
		if name == "" || names[name] {
			return nil, fmt.Errorf("auth: provider name %q is empty or used twice", name)
		}
		names[name] = true

		match := provider.Handles()
		if len(match.Issuers) == 0 && match.Prefix == "" {
			return nil, fmt.Errorf("auth: provider %q handles no credentials", name)
		}
		for _, issuer := range match.Issuers {
			if other, taken := authenticator.byIssuer[issuer]; taken {
				return nil, fmt.Errorf("auth: issuer %q is handled by both %q and %q", issuer, other.Name(), name)
			}
			authenticator.byIssuer[issuer] = provider
		}
		if match.Prefix != "" {
			for prefix, other := range authenticator.byPrefix {
				if strings.HasPrefix(match.Prefix, prefix) || strings.HasPrefix(prefix, match.Prefix) {
					return nil, fmt.Errorf("auth: prefix %q of %q overlaps %q of %q", match.Prefix, name, prefix, other.Name())
				}
			}
			authenticator.byPrefix[match.Prefix] = provider
		}
	}
	if len(names) == 0 {
		return nil, errors.New("auth: no identity providers configured")
	}
	return authenticator, nil
}

// Authenticate verifies credential with the provider that handles it.
// Every failure wraps ErrUnauthenticated.
func (authenticator *Authenticator) Authenticate(ctx context.Context, raw string) (*Principal, error) {
	if raw == "" {
		return nil, fmt.Errorf("%w: no credential", ErrUnauthenticated)
	}
	provider, credential, err := authenticator.route(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnauthenticated, err)
	}
	principal, err := provider.Authenticate(ctx, credential)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrUnauthenticated, provider.Name(), err)
	}
	principal.Provider = provider.Name()
	return principal, nil
}

func (authenticator *Authenticator) route(raw string) (IIdentityProvider, Credential, error) {
	for prefix, provider := range authenticator.byPrefix {
		if strings.HasPrefix(raw, prefix) {
			return provider, Credential{Raw: raw}, nil
		}
	}
	issuer, err := PeekIssuer(raw)
	if err != nil {
		return nil, Credential{}, err
	}
	provider, ok := authenticator.byIssuer[issuer]
	if !ok {
		return nil, Credential{}, fmt.Errorf("untrusted issuer %q", issuer)
	}
	return provider, Credential{Raw: raw, Issuer: issuer}, nil
}

// PeekIssuer reads the "iss" claim of a JWT without verifying anything. It
// only serves routing; providers verify the issuer themselves.
func PeekIssuer(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("not a jwt")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", errors.New("malformed jwt payload")
	}
	var claims struct {
		Issuer string `json:"iss"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Issuer == "" {
		return "", errors.New("jwt has no issuer")
	}
	return claims.Issuer, nil
}
