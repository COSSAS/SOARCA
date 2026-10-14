package entra

import (
	"fmt"
	"slices"
	"strings"

	"soarca/pkg/auth"
)

// claims are the Entra ID access token claims SOARCA uses.
type claims struct {
	TenantId          string            `json:"tid"`
	ObjectId          string            `json:"oid"`
	AuthorizedParty   string            `json:"azp"`   // calling app, v2 tokens
	AppId             string            `json:"appid"` // calling app, v1 tokens
	Scope             string            `json:"scp"`   // delegated scopes, user tokens only
	IdentityType      string            `json:"idtyp"` // "app" or "user", when the optional claim is configured
	Roles             []string          `json:"roles"`
	Groups            []string          `json:"groups"`
	ClaimNames        map[string]string `json:"_claim_names"` // set when groups overflow the token
	Name              string            `json:"name"`
	PreferredUsername string            `json:"preferred_username"`
	Upn               string            `json:"upn"`
}

func (tokenClaims claims) clientId() string {
	if tokenClaims.AuthorizedParty != "" {
		return tokenClaims.AuthorizedParty
	}
	return tokenClaims.AppId
}

func (tokenClaims claims) kind() string {
	switch tokenClaims.IdentityType {
	case auth.KindApp, auth.KindUser:
		return tokenClaims.IdentityType
	}
	// App-only tokens never carry delegated scopes.
	if tokenClaims.Scope == "" {
		return auth.KindApp
	}
	return auth.KindUser
}

// principal checks the Entra-specific claims of a token that passed
// signature, issuer, expiry and audience checks, and maps it to a Principal.
func (config Config) principal(tenantId string, token auth.Token, tokenClaims claims) (*auth.Principal, error) {
	if !strings.EqualFold(tokenClaims.TenantId, tenantId) {
		return nil, fmt.Errorf("tenant %q does not match issuer tenant %q", tokenClaims.TenantId, tenantId)
	}
	clientId := tokenClaims.clientId()
	if len(config.AllowedClientIds) > 0 && !slices.Contains(config.AllowedClientIds, clientId) {
		return nil, fmt.Errorf("client %q is not allowed", clientId)
	}
	kind := tokenClaims.kind()
	if kind == auth.KindUser && config.RequiredUserScope != "" &&
		!slices.Contains(strings.Fields(tokenClaims.Scope), config.RequiredUserScope) {
		return nil, fmt.Errorf("user token lacks scope %q", config.RequiredUserScope)
	}

	subject := tokenClaims.ObjectId
	if subject == "" {
		subject = token.Subject
	}
	namespace := "entra:" + strings.ToLower(tokenClaims.TenantId)
	self := auth.UserSubject(namespace, subject)
	if kind == auth.KindApp {
		// Administrators know apps by client ID, so access lists use it.
		self = auth.AppSubject(namespace, clientId)
	}
	subjects := []string{self}
	for _, group := range tokenClaims.Groups {
		subjects = append(subjects, auth.GroupSubject(namespace, strings.ToLower(group)))
	}
	_, groupsOverflowed := tokenClaims.ClaimNames["groups"]

	return &auth.Principal{
		Issuer:           token.Issuer,
		Subject:          subject,
		TenantId:         tokenClaims.TenantId,
		ClientId:         clientId,
		Kind:             kind,
		Name:             firstNonEmpty(tokenClaims.Name, tokenClaims.PreferredUsername, tokenClaims.Upn),
		Email:            firstNonEmpty(tokenClaims.PreferredUsername, tokenClaims.Upn),
		Roles:            tokenClaims.Roles,
		Permissions:      config.Roles.Permissions(tokenClaims.Roles),
		ExpiresAt:        token.Expiry,
		Subjects:         subjects,
		GroupsIncomplete: groupsOverflowed,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
