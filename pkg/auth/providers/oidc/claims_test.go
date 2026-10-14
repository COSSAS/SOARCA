package oidc

import (
	"testing"
	"time"

	"soarca/pkg/auth"
	"soarca/pkg/auth/roles"

	"github.com/go-playground/assert/v2"
)

var token = auth.Token{Issuer: "https://authentik.example/application/o/soarca/", Subject: "user-1", Expiry: time.Unix(2000000000, 0)}

func testConfig() Config {
	return Config{Name: "authentik", IssuerUrl: token.Issuer, Audiences: []string{"soarca"}}.withDefaults()
}

func TestGroupsGrantPermissions(t *testing.T) {
	principal := testConfig().principal(token, map[string]any{
		"groups": []any{"soarca_execute"}, "name": "Bob", "email": "bob@example.com", "azp": "soarca",
	})

	assert.Equal(t, principal.Name, "Bob")
	assert.Equal(t, principal.Email, "bob@example.com")
	assert.Equal(t, principal.ClientId, "soarca")
	assert.Equal(t, principal.Permissions, map[roles.Permission]bool{roles.PlaybookRead: true, roles.PlaybookExecute: true})
	assert.Equal(t, principal.ExpiresAt, token.Expiry)
}

func TestGauthAdminGroupKeepsFullAccess(t *testing.T) {
	principal := testConfig().principal(token, map[string]any{"groups": []any{DefaultAdminGroup}})

	for _, permission := range roles.All() {
		assert.Equal(t, principal.Has(permission), true)
	}
}

func TestSubjectsUseGroupIds(t *testing.T) {
	principal := testConfig().principal(token, map[string]any{
		"groups": []any{"soarca_read"}, "group_ids": []any{"uuid-1"},
	})

	assert.Equal(t, principal.Subjects, []string{"authentik:user:user-1", "authentik:group:uuid-1"})
}

func TestSubjectsFallBackToGroupNames(t *testing.T) {
	principal := testConfig().principal(token, map[string]any{"groups": []any{"soarca_read"}})

	assert.Equal(t, principal.Subjects, []string{"authentik:user:user-1", "authentik:group:soarca_read"})
}

func TestNameFallsBackToUsername(t *testing.T) {
	principal := testConfig().principal(token, map[string]any{"preferred_username": "bob"})

	assert.Equal(t, principal.Name, "bob")
}

func TestStringList(t *testing.T) {
	assert.Equal(t, stringList([]any{"a", 1, "b"}), []string{"a", "b"})
	assert.Equal(t, stringList("a b"), []string{"a", "b"})
	assert.Equal(t, len(stringList(nil)), 0)
}

func TestDefaults(t *testing.T) {
	config := Config{IssuerUrl: "https://idp/", Audiences: []string{"x"}}.withDefaults()

	assert.Equal(t, config.Name, "oidc")
	assert.Equal(t, config.GroupsClaim, "groups")
	assert.Equal(t, config.GroupIdsClaim, "group_ids")
	assert.NotEqual(t, config.Roles, nil)
}
