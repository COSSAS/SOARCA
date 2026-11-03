package auth

import (
	"time"

	"soarca/pkg/auth/roles"
)

// Kinds of principal.
const (
	KindUser = "user"
	KindApp  = "app"
)

// Principal is an authenticated caller, whichever way it authenticated.
type Principal struct {
	// Provider names the verifier that accepted the credential.
	Provider string `json:"provider"`
	Issuer   string `json:"issuer,omitempty"`
	// Subject is the provider's stable identifier for the caller.
	Subject  string `json:"subject"`
	TenantId string `json:"tenant_id,omitempty"`
	// ClientId is the calling application, when the provider reports one.
	ClientId string `json:"client_id,omitempty"`
	// Kind is KindUser, KindApp, or empty when the provider can't tell.
	Kind  string `json:"kind,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	// Roles holds the raw roles or groups from the credential.
	Roles       []string                  `json:"roles"`
	Permissions map[roles.Permission]bool `json:"permissions"`
	ExpiresAt   time.Time                 `json:"expires_at"`
	// Subjects lists what the caller is, for per-object access lists:
	// Subjects[0] is the caller itself, the rest are its groups. Build
	// entries with UserSubject, AppSubject and GroupSubject.
	Subjects []string `json:"subjects"`
	// GroupsIncomplete is set when the credential could not list every
	// group, so group-based grants may be missing.
	GroupsIncomplete bool `json:"groups_incomplete,omitempty"`
}

// Has reports whether the principal holds permission. roles.Admin implies
// every permission, whichever provider granted it.
func (principal *Principal) Has(permission roles.Permission) bool {
	return principal != nil && (principal.Permissions[permission] || principal.Permissions[roles.Admin])
}

// Self returns the subject that identifies the caller itself.
func (principal *Principal) Self() string {
	if principal == nil || len(principal.Subjects) == 0 {
		return ""
	}
	return principal.Subjects[0]
}

// UserSubject, AppSubject and GroupSubject build namespaced subjects of the
// form "<namespace>:<kind>:<id>", so identifiers from different providers
// never collide. Providers choose their namespace.
func UserSubject(namespace string, id string) string {
	return namespace + ":user:" + id
}

func AppSubject(namespace string, id string) string {
	return namespace + ":app:" + id
}

func GroupSubject(namespace string, id string) string {
	return namespace + ":group:" + id
}
