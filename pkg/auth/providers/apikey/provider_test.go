package apikey

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"soarca/pkg/auth"
	"soarca/pkg/auth/roles"

	"github.com/go-playground/assert/v2"
)

func writeKeys(t *testing.T, keys []Key) string {
	content, err := json.Marshal(keys)
	assert.Equal(t, err, nil)
	path := filepath.Join(t.TempDir(), "keys.json")
	assert.Equal(t, os.WriteFile(path, content, 0o600), nil)
	return path
}

func TestGenerate(t *testing.T) {
	plaintext, hash, err := Generate()

	assert.Equal(t, err, nil)
	assert.Equal(t, strings.HasPrefix(plaintext, Prefix), true)
	assert.Equal(t, hash, Hash(plaintext))
}

func TestVerifyValidKey(t *testing.T) {
	plaintext, hash, _ := Generate()
	provider := New(MemoryStore{hash: {Name: "webhook", Sha256: hash, Permissions: []roles.Permission{roles.PlaybookExecute}, Expires: time.Now().Add(time.Hour)}})

	principal, err := provider.Authenticate(context.Background(), auth.Credential{Raw: plaintext})

	assert.Equal(t, err, nil)
	assert.Equal(t, principal.Self(), "apikey:app:webhook")
	assert.Equal(t, principal.Permissions, map[roles.Permission]bool{roles.PlaybookExecute: true})
	assert.Equal(t, provider.Name(), "apikey")
	assert.Equal(t, provider.Handles(), auth.Match{Prefix: Prefix})
}

func TestVerifyRejectsExpiredAndUnknownKeys(t *testing.T) {
	expired, hash, _ := Generate()
	provider := New(MemoryStore{hash: {Name: "old", Sha256: hash, Expires: time.Now().Add(-time.Hour)}})

	_, err := provider.Authenticate(context.Background(), auth.Credential{Raw: expired})
	assert.NotEqual(t, err, nil)

	_, err = provider.Authenticate(context.Background(), auth.Credential{Raw: Prefix + "unknown"})
	assert.NotEqual(t, err, nil)
}

func TestLoadFile(t *testing.T) {
	_, hash, _ := Generate()
	expires := time.Now().Add(time.Hour).UTC().Truncate(time.Second)

	store, err := LoadFile(writeKeys(t, []Key{{Name: "webhook", Sha256: hash, Permissions: []roles.Permission{roles.PlaybookRead}, Expires: expires}}))

	assert.Equal(t, err, nil)
	key, ok := store.Lookup(hash)
	assert.Equal(t, ok, true)
	assert.Equal(t, key.Name, "webhook")
	assert.Equal(t, key.Expires.Equal(expires), true)
}

func TestLoadFileRejectsInvalidKeys(t *testing.T) {
	_, hash, _ := Generate()
	for name, key := range map[string]Key{
		"no name":   {Sha256: hash, Expires: time.Now()},
		"bad hash":  {Name: "x", Sha256: "abc", Expires: time.Now()},
		"no expiry": {Name: "x", Sha256: hash},
	} {
		if _, err := LoadFile(writeKeys(t, []Key{key})); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}

	_, err := LoadFile(filepath.Join(t.TempDir(), "missing.json"))
	assert.NotEqual(t, err, nil)
}
