package entra

import "soarca/pkg/auth/roles"

// DefaultRoles maps the app-role values defined on the SOARCA API app
// registration in Entra ID to SOARCA permissions.
func DefaultRoles() roles.RoleMap {
	return roles.RoleMap{
		"Playbook.Read":    {roles.PlaybookRead},
		"Playbook.Write":   {roles.PlaybookRead, roles.PlaybookWrite},
		"Playbook.Execute": {roles.PlaybookRead, roles.PlaybookExecute},
		"SOARCA.Admin":     roles.All(),
	}
}
