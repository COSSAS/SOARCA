package auth

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OidsConfig struct {
	OIDCIssuer string
	AuthURL    string
	OAuth2     *oauth2.Config
	UserConfig *UserConfig
	provider   *oidc.Provider
}
type Authenticator struct {
	config *OidsConfig
}

type UserConfig struct {
	AuthBlockNew     bool
	AuthDefaultRoles []string
	AuthAdminUsers   []string

	OIDCClaimUsername string
	OIDCClaimEmail    string
	OIDCClaimName     string
	OIDCClaimGroups   string
}

func NewAuthenticatator(ctx context.Context, config *OidsConfig) (*Authenticator, error) {
	authenticator := &Authenticator{
		config: config,
	}
	err := authenticator.Load(ctx)
	if err != nil {
		return nil, err
	}
	return authenticator, nil
}

func (a *Authenticator) Load(ctx context.Context) error {
	provider, err := oidc.NewProvider(ctx, a.config.OIDCIssuer)
	if err == nil {
		a.config.provider = provider
		a.config.OAuth2.Endpoint = provider.Endpoint()
		if a.config.AuthURL != "" {
			a.config.OAuth2.Endpoint.AuthURL = a.config.AuthURL
		}
	}

	return err
}

func (a *Authenticator) Verifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	if a.config.provider == nil {
		if err := a.Load(ctx); err != nil {
			return nil, err
		}
	}

	config := &oidc.Config{ClientID: a.config.OAuth2.ClientID}
	if a.config.AuthURL != "" {
		config.SkipIssuerCheck = true
	}

	return a.config.provider.Verifier(config), nil
}
