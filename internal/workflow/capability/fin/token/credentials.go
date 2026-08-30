// Package token generates and hashes the bearer credentials used by the
// Fin protocol (fin_token, and the registration token comparison). Tokens
// are high-entropy random secrets; SOARCA never persists one in plaintext
// (see pkg/models/fin.Record.FinTokenHash) - only its SHA-256 hash, which
// is what callers compare/look up against on every authenticated call.
package token

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
)

// tokenBytes is the amount of random entropy in a generated token, before
// hex-encoding (32 bytes = 256 bits).
const tokenBytes = 32

// Generate returns a new, high-entropy random token suitable for use as a
// fin_token.
func Generate() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", errors.New("failed to generate random token: " + err.Error())
	}
	return hex.EncodeToString(buf), nil
}

// Hash returns the SHA-256 hash (hex-encoded) of token, for storage/
// lookup/comparison instead of the plaintext value. No per-token salt is
// used: token itself is already a high-entropy random secret (see
// Generate), so unlike a user password hash, there is no risk of a
// precomputed dictionary/rainbow-table attack to defend against.
func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Equal does a constant-time comparison of two tokens (e.g. a presented
// registration token against the configured one), to avoid leaking timing
// information about how much of a prefix matched.
func Equal(a string, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
