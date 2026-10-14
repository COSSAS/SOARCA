// Package entra verifies access tokens issued by Microsoft Entra ID for the
// SOARCA API app registration.
package entra

import (
	"errors"
	"strings"

	"soarca/pkg/auth/roles"
)

// DefaultAuthority is the Entra ID login host for the public cloud.
const DefaultAuthority = "https://login.microsoftonline.com"

const v1IssuerBase = "https://sts.windows.net"

// Config describes the app registration that represents the SOARCA API
// ("Expose an API"). Callers must request tokens for this API; tokens for
// Microsoft Graph cannot be verified by anyone but Microsoft.
type Config struct {
	// TenantIds that may call SOARCA.
	TenantIds []string
	// ApiClientId is the application (client) ID of the SOARCA API app
	// registration. v2 access tokens carry it as their audience.
	ApiClientId string
	// ApiAppIdUri is the Application ID URI, e.g. "api://soarca". v1 access
	// tokens carry it as their audience. Optional.
	ApiAppIdUri string
	// AllowedClientIds restricts which applications may call SOARCA.
	// Empty allows any client in an allowed tenant.
	AllowedClientIds []string
	// RequiredUserScope is the delegated scope user tokens must carry,
	// e.g. "access_as_user". Empty skips the check.
	RequiredUserScope string
	// AcceptV1Tokens also accepts tokens from https://sts.windows.net/<tenant>/.
	AcceptV1Tokens bool
	// Authority is the login host. Empty means DefaultAuthority.
	Authority string
	// Roles maps app-role values to permissions. Nil means
	// DefaultRoles().
	Roles roles.RoleMap
}

func (config Config) withDefaults() Config {
	if config.Authority == "" {
		config.Authority = DefaultAuthority
	}
	config.Authority = strings.TrimRight(config.Authority, "/")
	if config.Roles == nil {
		config.Roles = DefaultRoles()
	}
	return config
}

func (config Config) validate() error {
	if len(config.TenantIds) == 0 {
		return errors.New("entra: at least one tenant ID is required")
	}
	if config.ApiClientId == "" {
		return errors.New("entra: the SOARCA API client ID is required")
	}
	return nil
}
