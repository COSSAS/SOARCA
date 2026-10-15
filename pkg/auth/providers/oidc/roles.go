package oidc

import "soarca/pkg/auth/roles"

// DefaultAdminGroup is the group that older SOARCA versions required for all
// access. It keeps full access.
const DefaultAdminGroup = "soarca_admin"

// DefaultRoles maps group names to SOARCA permissions. adminGroup gets every
// permission; empty means DefaultAdminGroup.
func DefaultRoles(adminGroup string) roles.RoleMap {
	if adminGroup == "" {
		adminGroup = DefaultAdminGroup
	}
	return roles.RoleMap{
		"soarca_read":    {roles.PlaybookRead},
		"soarca_write":   {roles.PlaybookRead, roles.PlaybookWrite},
		"soarca_execute": {roles.PlaybookRead, roles.PlaybookExecute},
		adminGroup:       roles.All(),
	}
}
