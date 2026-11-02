package api

import "soarca/pkg/auth"

// ProtectedResourceMetadata describes SOARCA as an OAuth protected resource
// (RFC 9728). Clients read it to learn which identity providers SOARCA
// accepts tokens from.
type ProtectedResourceMetadata struct {
	// Resource is SOARCA's public URL.
	Resource     string `json:"resource"`
	ResourceName string `json:"resource_name"`
	// AuthorizationServers lists the issuer of every identity provider
	// SOARCA accepts tokens from. Empty when authentication is disabled.
	AuthorizationServers   []string `json:"authorization_servers"`
	ScopesSupported        []string `json:"scopes_supported,omitempty"`
	BearerMethodsSupported []string `json:"bearer_methods_supported"`
}

type AuthUser struct {
	Principal *auth.Principal `json:"principal"`
}
