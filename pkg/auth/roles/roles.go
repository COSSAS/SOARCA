// Package roles defines SOARCA's permissions and the RoleMap type that
// identity providers use to grant them. Which roles or groups grant which
// permissions is up to each provider.
package roles

// Permission is something a caller may do in SOARCA. Routes require
// permissions; identity providers grant them through a RoleMap.
type Permission string

const (
	PlaybookRead    Permission = "playbook:read"
	PlaybookWrite   Permission = "playbook:write"
	PlaybookExecute Permission = "playbook:execute"
	Admin           Permission = "soarca:admin"
)

// All returns every permission.
func All() []Permission {
	return []Permission{PlaybookRead, PlaybookWrite, PlaybookExecute, Admin}
}

// RoleMap maps a role or group name from an identity provider to the
// permissions it grants.
type RoleMap map[string][]Permission

// Permissions returns the union of the permissions granted by roles.
// Unknown roles grant nothing.
func (roleMap RoleMap) Permissions(roles []string) map[Permission]bool {
	granted := map[Permission]bool{}
	for _, role := range roles {
		for _, permission := range roleMap[role] {
			granted[permission] = true
		}
	}
	return granted
}
