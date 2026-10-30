package controller

import (
	"fmt"
	"strings"

	"soarca/internal/config"
	"soarca/pkg/auth"
	"soarca/pkg/auth/providers/apikey"
	"soarca/pkg/auth/providers/entra"
	"soarca/pkg/auth/providers/oidc"
	"soarca/pkg/models/api"
)

// newAuthenticator builds the authenticator for every provider that
// authConfig enables. config.Load has already checked that each enabled
// provider is complete and that at least one is enabled.
func newAuthenticator(authConfig config.Auth) (auth.IAuthenticator, error) {
	if !authConfig.Enabled {
		log.Warning("authentication is disabled: every API request is accepted with full permissions")
		return auth.Anonymous{}, nil
	}

	var providers []auth.IIdentityProvider
	if authConfig.OidcEnabled() {
		provider, err := oidc.New(oidc.Config{
			IssuerUrl:     authConfig.Issuer,
			Audiences:     []string{authConfig.ClientId},
			Roles:         oidc.DefaultRoles(authConfig.AdminGroup),
			SkipTlsVerify: authConfig.SkipTlsVerify,
		})
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
		log.Info("accepting OIDC tokens from ", authConfig.Issuer)
	}

	if authConfig.EntraEnabled() {
		entraConfig := authConfig.Entra
		provider, err := entra.New(entra.Config{
			TenantIds:         entraConfig.TenantIds,
			ApiClientId:       entraConfig.ApiClientId,
			ApiAppIdUri:       entraConfig.ApiAppIdUri,
			AllowedClientIds:  entraConfig.AllowedClientIds,
			RequiredUserScope: entraConfig.RequiredUserScope,
			AcceptV1Tokens:    entraConfig.AcceptV1Tokens,
			Authority:         entraConfig.Authority,
		})
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
		log.Info("accepting Entra ID tokens from tenants ", entraConfig.TenantIds)
	}

	if authConfig.ApiKeysFile != "" {
		store, err := apikey.LoadFile(authConfig.ApiKeysFile)
		if err != nil {
			return nil, err
		}
		providers = append(providers, apikey.New(store))
		log.Info("accepting ", len(store), " API keys from ", authConfig.ApiKeysFile)
	}

	return auth.New(providers)
}

// newAuthConfig lists the providers users can sign in with.
func newAuthConfig(authConfig config.Auth) api.AuthConfig {
	published := api.AuthConfig{AuthenticationEnabled: authConfig.Enabled, Providers: []api.AuthProvider{}}
	if !authConfig.Enabled {
		return published
	}
	if authConfig.OidcEnabled() {
		published.Providers = append(published.Providers, api.AuthProvider{
			Id:       "oidc",
			Issuer:   authConfig.Issuer,
			ClientId: authConfig.ClientId,
			Scopes:   []string{},
		})
	}
	entraConfig := authConfig.Entra
	if authConfig.EntraEnabled() && entraConfig.PublicClientId != "" {
		for _, tenantId := range entraConfig.TenantIds {
			published.Providers = append(published.Providers, api.AuthProvider{
				Id:       entra.Name,
				Issuer:   fmt.Sprintf("%s/%s/v2.0", strings.TrimRight(entraConfig.Authority, "/"), tenantId),
				ClientId: entraConfig.PublicClientId,
				Scopes:   []string{entraConfig.UserScope()},
			})
		}
	}
	return published
}
