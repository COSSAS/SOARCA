package oidc

import (
	"testing"

	"soarca/pkg/auth/roles"

	"github.com/go-playground/assert/v2"
)

func TestDefaultRolesKeepGauthAdminGroup(t *testing.T) {
	assert.Equal(t, len(DefaultRoles("").Permissions([]string{DefaultAdminGroup})), len(roles.All()))
}

func TestDefaultRolesWithCustomAdminGroup(t *testing.T) {
	defaults := DefaultRoles("soc_leads")

	assert.Equal(t, len(defaults.Permissions([]string{"soc_leads"})), len(roles.All()))
	assert.Equal(t, len(defaults.Permissions([]string{DefaultAdminGroup})), 0)
}

func TestDefaultRolesForPlaybookGroups(t *testing.T) {
	defaults := DefaultRoles("")

	assert.Equal(t, defaults.Permissions([]string{"soarca_read"}), map[roles.Permission]bool{roles.PlaybookRead: true})
	assert.Equal(t, defaults.Permissions([]string{"soarca_write"}), map[roles.Permission]bool{roles.PlaybookRead: true, roles.PlaybookWrite: true})
	assert.Equal(t, defaults.Permissions([]string{"soarca_execute"}), map[roles.Permission]bool{roles.PlaybookRead: true, roles.PlaybookExecute: true})
}
