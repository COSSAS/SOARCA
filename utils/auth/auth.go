package auth

import (
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type AuthConfig struct {
	InitialUser     string
	InitialPassword string
	InitialAPIKey   string

	OIDCIssuer       string
	AuthURL          string
	OAuth2           *oauth2.Config
	UserCreateConfig *UserCreateConfig

	provider *oidc.Provider
}
