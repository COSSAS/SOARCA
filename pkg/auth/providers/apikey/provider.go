// Package apikey verifies static API keys, for callers that cannot request
// an OAuth token, such as a webhook source that sends a fixed header.
// Only the SHA-256 of each key is stored.
package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"soarca/pkg/auth"
	"soarca/pkg/auth/roles"
)

// Prefix marks SOARCA API keys, so the authenticator can route them and
// secret scanners can recognise leaked ones.
const Prefix = "soarca_"

// Name labels this provider in principals and logs.
const Name = "apikey"

// Key is a stored API key.
type Key struct {
	Name        string             `json:"name"`
	Sha256      string             `json:"sha256"`
	Permissions []roles.Permission `json:"permissions"`
	Expires     time.Time          `json:"expires"`
}

func (key Key) validate() error {
	if key.Name == "" {
		return errors.New("api key needs a name")
	}
	if len(key.Sha256) != sha256.Size*2 {
		return fmt.Errorf("api key %q needs a 64-character sha256", key.Name)
	}
	if key.Expires.IsZero() {
		return fmt.Errorf("api key %q has no expiry", key.Name)
	}
	return nil
}

// IStore looks up a key by the hex SHA-256 of its plaintext.
type IStore interface {
	Lookup(sha256Hex string) (Key, bool)
}

// MemoryStore is a read-only store, usually loaded with LoadFile.
type MemoryStore map[string]Key

func (store MemoryStore) Lookup(sha256Hex string) (Key, bool) {
	key, ok := store[sha256Hex]
	return key, ok
}

// LoadFile reads a JSON array of Key.
func LoadFile(path string) (MemoryStore, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("apikey: %w", err)
	}
	var keys []Key
	if err := json.Unmarshal(content, &keys); err != nil {
		return nil, fmt.Errorf("apikey: %w", err)
	}
	store := MemoryStore{}
	for _, key := range keys {
		if err := key.validate(); err != nil {
			return nil, fmt.Errorf("apikey: %w", err)
		}
		store[key.Sha256] = key
	}
	return store, nil
}

// Generate returns a new plaintext key, to hand to the caller once, and the
// hash to store.
func Generate() (plaintext string, sha256Hex string, err error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", "", err
	}
	plaintext = Prefix + base64.RawURLEncoding.EncodeToString(random)
	return plaintext, Hash(plaintext), nil
}

// Hash returns the hex SHA-256 of a plaintext key.
func Hash(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// Provider checks API keys against a store.
type Provider struct {
	store IStore
}

func New(store IStore) *Provider {
	return &Provider{store: store}
}

func (provider *Provider) Name() string {
	return Name
}

// Handles declares the API key prefix.
func (provider *Provider) Handles() auth.Match {
	return auth.Match{Prefix: Prefix}
}

// Authenticate looks up the key and maps it to a Principal.
func (provider *Provider) Authenticate(_ context.Context, credential auth.Credential) (*auth.Principal, error) {
	// Looking up by hash leaves no timing signal about the secret itself.
	key, ok := provider.store.Lookup(Hash(credential.Raw))
	if !ok {
		return nil, errors.New("unknown api key")
	}
	if time.Now().After(key.Expires) {
		return nil, fmt.Errorf("api key %q expired", key.Name)
	}
	permissions := map[roles.Permission]bool{}
	names := make([]string, 0, len(key.Permissions))
	for _, permission := range key.Permissions {
		permissions[permission] = true
		names = append(names, string(permission))
	}
	return &auth.Principal{
		Subject:     key.Name,
		Kind:        auth.KindApp,
		Name:        key.Name,
		Roles:       names,
		Permissions: permissions,
		ExpiresAt:   key.Expires,
		Subjects:    []string{auth.AppSubject(Name, key.Name)},
	}, nil
}
