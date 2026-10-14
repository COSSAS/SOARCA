package oidc

import (
	"strings"

	"soarca/pkg/auth"
)

// principal maps the claims of a token that passed signature, issuer,
// expiry and audience checks to a Principal.
func (config Config) principal(token auth.Token, claims map[string]any) *auth.Principal {
	groups := stringList(claims[config.GroupsClaim])

	// Access lists key on group IDs, so renaming a group doesn't break them.
	groupIds := groups
	if raw, ok := claims[config.GroupIdsClaim]; ok {
		groupIds = stringList(raw)
	}
	subjects := []string{auth.UserSubject(config.Name, token.Subject)}
	for _, group := range groupIds {
		subjects = append(subjects, auth.GroupSubject(config.Name, group))
	}

	name := stringClaim(claims["name"])
	if name == "" {
		name = stringClaim(claims["preferred_username"])
	}
	return &auth.Principal{
		Issuer:      token.Issuer,
		Subject:     token.Subject,
		ClientId:    stringClaim(claims["azp"]),
		Name:        name,
		Email:       stringClaim(claims["email"]),
		Roles:       groups,
		Permissions: config.Roles.Permissions(groups),
		ExpiresAt:   token.Expiry,
		Subjects:    subjects,
	}
}

// stringList reads a claim that is either a JSON array of strings or a
// space-separated string.
func stringList(value any) []string {
	switch typed := value.(type) {
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				out = append(out, text)
			}
		}
		return out
	case string:
		return strings.Fields(typed)
	}
	return nil
}

func stringClaim(value any) string {
	text, _ := value.(string)
	return text
}
