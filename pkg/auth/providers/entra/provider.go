package entra

import (
	"context"
	"fmt"
	"sort"

	"soarca/pkg/auth"

	"github.com/coreos/go-oidc/v3/oidc"
)

// Name labels this provider in principals and logs.
const Name = "entra"

// Provider verifies Entra ID access tokens for the configured tenants.
type Provider struct {
	config  Config
	tenants map[string]tenantValidator // keyed by issuer
}

type tenantValidator struct {
	tenantId  string
	validator *auth.TokenValidator
}

// New builds a Provider. It doesn't contact Entra ID; signing keys are
// fetched on first use and cached.
func New(config Config) (*Provider, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	config = config.withDefaults()
	audiences := []string{config.ApiClientId, config.ApiAppIdUri}

	provider := &Provider{config: config, tenants: map[string]tenantValidator{}}
	for _, tenantId := range config.TenantIds {
		// go-oidc keeps this context for later key refreshes, so it must
		// outlive any request.
		keys := oidc.NewRemoteKeySet(context.Background(), fmt.Sprintf("%s/%s/discovery/v2.0/keys", config.Authority, tenantId))
		issuers := []string{fmt.Sprintf("%s/%s/v2.0", config.Authority, tenantId)}
		if config.AcceptV1Tokens {
			issuers = append(issuers, fmt.Sprintf("%s/%s/", v1IssuerBase, tenantId))
		}
		for _, issuer := range issuers {
			provider.tenants[issuer] = tenantValidator{
				tenantId:  tenantId,
				validator: auth.NewTokenValidator(issuer, keys, audiences, []string{oidc.RS256}),
			}
		}
	}
	return provider, nil
}

func (provider *Provider) Name() string {
	return Name
}

// Handles declares the issuer of every allowed tenant.
func (provider *Provider) Handles() auth.Match {
	issuers := make([]string, 0, len(provider.tenants))
	for issuer := range provider.tenants {
		issuers = append(issuers, issuer)
	}
	sort.Strings(issuers)
	return auth.Match{Issuers: issuers}
}

// Authenticate verifies the token and maps it to a Principal.
func (provider *Provider) Authenticate(ctx context.Context, credential auth.Credential) (*auth.Principal, error) {
	tenant, ok := provider.tenants[credential.Issuer]
	if !ok {
		return nil, fmt.Errorf("issuer %q is not an allowed tenant", credential.Issuer)
	}
	validated, err := tenant.validator.Validate(ctx, credential.Raw)
	if err != nil {
		return nil, err
	}
	var tokenClaims claims
	if err := validated.Claims(&tokenClaims); err != nil {
		return nil, err
	}
	return provider.config.principal(tenant.tenantId, validated, tokenClaims)
}
