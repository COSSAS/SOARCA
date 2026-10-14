package entra

import (
	"testing"

	"soarca/pkg/auth/roles"

	"github.com/go-playground/assert/v2"
)

func TestDefaultRoles(t *testing.T) {
	defaults := DefaultRoles()

	assert.Equal(t, defaults.Permissions([]string{"Playbook.Read"}), map[roles.Permission]bool{roles.PlaybookRead: true})
	assert.Equal(t, defaults.Permissions([]string{"Playbook.Write"}), map[roles.Permission]bool{roles.PlaybookRead: true, roles.PlaybookWrite: true})
	assert.Equal(t, defaults.Permissions([]string{"Playbook.Execute"}), map[roles.Permission]bool{roles.PlaybookRead: true, roles.PlaybookExecute: true})
	assert.Equal(t, len(defaults.Permissions([]string{"SOARCA.Admin"})), len(roles.All()))
}

func TestDefaultRolesAreCaseSensitive(t *testing.T) {
	// Entra sends app-role values exactly as defined on the registration.
	assert.Equal(t, len(DefaultRoles().Permissions([]string{"playbook.read"})), 0)
}
