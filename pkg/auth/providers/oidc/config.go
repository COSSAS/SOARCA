// Package oidc verifies access tokens from a standard OpenID Connect
// provider, such as authentik, found through discovery.
package oidc

import (
	"errors"

	"soarca/pkg/auth/roles"
)

// Config describes one OIDC provider.
type Config struct {
	// Name labels the provider in principals and access-list subjects.
	// Empty means "oidc".
	Name string
	// IssuerUrl must equal the tokens' "iss" claim exactly. For authentik:
	// https://<host>/application/o/<slug>/ including the trailing slash.
	IssuerUrl string
	// Audiences accepted in "aud", normally the provider's client ID.
	Audiences []string
	// GroupsClaim holds the group names used for permissions.
	// Empty means "groups".
	GroupsClaim string
	// GroupIdsClaim holds stable group IDs for access lists. Empty means
	// "group_ids". Tokens without it fall back to the group names.
	GroupIdsClaim string
	// Roles maps group names to permissions. Nil means
	// DefaultRoles(DefaultAdminGroup).
	Roles roles.RoleMap
	// SkipTlsVerify disables certificate checks towards the provider. Only
	// for test setups with self-signed certificates.
	SkipTlsVerify bool
}

func (config Config) withDefaults() Config {
	if config.Name == "" {
		config.Name = "oidc"
	}
	if config.GroupsClaim == "" {
		config.GroupsClaim = "groups"
	}
	if config.GroupIdsClaim == "" {
		config.GroupIdsClaim = "group_ids"
	}
	if config.Roles == nil {
		config.Roles = DefaultRoles(DefaultAdminGroup)
	}
	return config
}

func (config Config) validate() error {
	if config.IssuerUrl == "" {
		return errors.New("oidc: the issuer URL is required")
	}
	if len(config.Audiences) == 0 {
		return errors.New("oidc: at least one audience (client ID) is required")
	}
	return nil
}
