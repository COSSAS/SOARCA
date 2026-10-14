package auth

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

// ClockLeeway is how far past expiry a token is still accepted, to absorb
// clock differences between SOARCA and the identity provider.
const ClockLeeway = 30 * time.Second

// TokenValidator checks a JWT's signature, issuer, expiry and audience.
// Provider packages use it and add their own claim checks on top.
type TokenValidator struct {
	audiences []string
	verifier  *oidc.IDTokenVerifier
}

// Token is a validated JWT.
type Token struct {
	Issuer   string
	Subject  string
	Audience []string
	Expiry   time.Time
	verified *oidc.IDToken
}

// Claims decodes the token's claims into target.
func (token Token) Claims(target any) error {
	return token.verified.Claims(target)
}

// NewTokenValidator accepts tokens from issuer, signed by a key in keys with
// one of algorithms (RS256 when empty), carrying one of audiences.
func NewTokenValidator(issuer string, keys oidc.KeySet, audiences []string, algorithms []string) *TokenValidator {
	return &TokenValidator{
		audiences: audiences,
		verifier: oidc.NewVerifier(issuer, keys, &oidc.Config{
			// go-oidc checks a single client ID; audiences are checked below.
			SkipClientIDCheck:    true,
			SupportedSigningAlgs: algorithms,
			Now:                  func() time.Time { return time.Now().Add(-ClockLeeway) },
		}),
	}
}

// Validate verifies rawToken and returns it.
func (validator *TokenValidator) Validate(ctx context.Context, rawToken string) (Token, error) {
	verified, err := validator.verifier.Verify(ctx, rawToken)
	if err != nil {
		return Token{}, err
	}
	if !containsAny(verified.Audience, validator.audiences) {
		return Token{}, fmt.Errorf("audience %v is not accepted", verified.Audience)
	}
	return Token{
		Issuer:   verified.Issuer,
		Subject:  verified.Subject,
		Audience: verified.Audience,
		Expiry:   verified.Expiry,
		verified: verified,
	}, nil
}

func containsAny(have []string, want []string) bool {
	for _, candidate := range want {
		if candidate != "" && slices.Contains(have, candidate) {
			return true
		}
	}
	return false
}
