// Package mock_identity_provider is an in-process identity provider for
// tests that need real signed JWTs: it serves a JWKS and OIDC discovery
// documents and signs tokens with its key.
package fake_identity_provider

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jose "github.com/go-jose/go-jose/v4"
)

const keyId = "test-key"

type IdentityProvider struct {
	Server  *httptest.Server
	key     *rsa.PrivateKey
	unknown *rsa.PrivateKey
	mux     *http.ServeMux
}

// New starts a provider that is shut down when the test ends. With useTls
// it serves HTTPS with a self-signed certificate.
func New(t *testing.T, useTls bool) *IdentityProvider {
	t.Helper()
	provider := &IdentityProvider{
		key:     generateKey(t),
		unknown: generateKey(t),
		mux:     http.NewServeMux(),
	}
	if useTls {
		provider.Server = httptest.NewTLSServer(provider.mux)
	} else {
		provider.Server = httptest.NewServer(provider.mux)
	}
	t.Cleanup(provider.Server.Close)
	return provider
}

// Url returns the absolute URL of path on this provider.
func (provider *IdentityProvider) Url(path string) string {
	return provider.Server.URL + path
}

// ServeKeys publishes the signing key as a JWKS at path.
func (provider *IdentityProvider) ServeKeys(path string) {
	keySet := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
		{Key: &provider.key.PublicKey, KeyID: keyId, Algorithm: string(jose.RS256), Use: "sig"},
	}}
	provider.mux.HandleFunc(path, func(writer http.ResponseWriter, _ *http.Request) {
		writeJson(writer, keySet)
	})
}

// ServeDiscovery publishes an OIDC discovery document for the issuer at
// issuerPath (ending in "/"), with its JWKS next to it.
func (provider *IdentityProvider) ServeDiscovery(issuerPath string) {
	provider.ServeKeys(issuerPath + "jwks/")
	provider.mux.HandleFunc(issuerPath+".well-known/openid-configuration", func(writer http.ResponseWriter, _ *http.Request) {
		writeJson(writer, map[string]any{
			"issuer":                                provider.Url(issuerPath),
			"jwks_uri":                              provider.Url(issuerPath + "jwks/"),
			"authorization_endpoint":                provider.Url(issuerPath + "authorize/"),
			"token_endpoint":                        provider.Url(issuerPath + "token/"),
			"id_token_signing_alg_values_supported": []string{string(jose.RS256)},
		})
	})
}

// Sign returns a JWT with claims, signed with the published key. It adds
// "exp" one hour ahead unless claims sets it.
func (provider *IdentityProvider) Sign(t *testing.T, claims map[string]any) string {
	return sign(t, provider.key, claims)
}

// SignWithUnknownKey signs with a key that is never published.
func (provider *IdentityProvider) SignWithUnknownKey(t *testing.T, claims map[string]any) string {
	return sign(t, provider.unknown, claims)
}

func sign(t *testing.T, key *rsa.PrivateKey, claims map[string]any) string {
	t.Helper()
	payload := map[string]any{"iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix()}
	for name, value := range claims {
		payload[name] = value
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", keyId))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	signed, err := signer.Sign(encoded)
	if err != nil {
		t.Fatal(err)
	}
	compact, err := signed.CompactSerialize()
	if err != nil {
		t.Fatal(err)
	}
	return compact
}

func generateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func writeJson(writer http.ResponseWriter, value any) {
	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode(value)
}
