package api

import "soarca/pkg/auth"

// AuthProvider is an identity provider users can sign in with, using the
// authorization code flow with PKCE. ClientId is a public client; there is
// no secret.
type AuthProvider struct {
	// Id matches the provider of the resulting principal.
	Id       string   `json:"id"`
	Issuer   string   `json:"issuer"`
	ClientId string   `json:"client_id"`
	Scopes   []string `json:"scopes"`
}

type AuthConfig struct {
	AuthenticationEnabled bool           `json:"authentication_enabled"`
	Providers             []AuthProvider `json:"providers"`
}

type AuthUser struct {
	Principal *auth.Principal `json:"principal"`
}
