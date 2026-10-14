package oidc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"soarca/internal/logger"
	"soarca/pkg/auth"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
)

const discoveryTimeout = 10 * time.Second

var log *logger.Log

func init() {
	log = logger.Logger(reflect.TypeOf(Provider{}).PkgPath(), logger.Info, "", logger.Json)
}

// Provider verifies tokens from one OIDC provider. Discovery runs on the
// first token and is retried after a failure, so SOARCA can start before
// the provider is reachable.
type Provider struct {
	config Config
	client *http.Client

	mutex     sync.Mutex
	validator *auth.TokenValidator
}

// New builds a Provider without contacting the identity provider.
func New(config Config) (*Provider, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	config = config.withDefaults()
	client := http.DefaultClient
	if config.SkipTlsVerify {
		log.Warning(fmt.Sprintf("TLS verification towards %s is disabled; use this for testing only", config.IssuerUrl))
		client = &http.Client{Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // explicitly configured with SkipTlsVerify
		}}
	}
	return &Provider{config: config, client: client}, nil
}

func (provider *Provider) Name() string {
	return provider.config.Name
}

// Handles declares the configured issuer.
func (provider *Provider) Handles() auth.Match {
	return auth.Match{Issuers: []string{provider.config.IssuerUrl}}
}

// Authenticate verifies the token and maps it to a Principal.
func (provider *Provider) Authenticate(ctx context.Context, credential auth.Credential) (*auth.Principal, error) {
	validator, err := provider.tokenValidator(ctx)
	if err != nil {
		return nil, err
	}
	validated, err := validator.Validate(ctx, credential.Raw)
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := validated.Claims(&claims); err != nil {
		return nil, err
	}
	return provider.config.principal(validated, claims), nil
}

func (provider *Provider) tokenValidator(ctx context.Context) (*auth.TokenValidator, error) {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()
	if provider.validator != nil {
		return provider.validator, nil
	}

	discoveryCtx, cancel := context.WithTimeout(gooidc.ClientContext(ctx, provider.client), discoveryTimeout)
	defer cancel()
	discovered, err := gooidc.NewProvider(discoveryCtx, provider.config.IssuerUrl)
	if err != nil {
		return nil, fmt.Errorf("discovery: %w", err)
	}
	var metadata struct {
		JwksUri    string   `json:"jwks_uri"`
		Algorithms []string `json:"id_token_signing_alg_values_supported"`
	}
	if err := discovered.Claims(&metadata); err != nil || metadata.JwksUri == "" {
		return nil, fmt.Errorf("discovery: %s has no jwks_uri", provider.config.IssuerUrl)
	}

	// The key set outlives this request, so it gets its own context.
	keys := gooidc.NewRemoteKeySet(gooidc.ClientContext(context.Background(), provider.client), metadata.JwksUri)
	provider.validator = auth.NewTokenValidator(provider.config.IssuerUrl, keys, provider.config.Audiences, metadata.Algorithms)
	return provider.validator, nil
}
