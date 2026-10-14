package entra

import (
	"testing"
	"time"

	"soarca/pkg/auth"
	"soarca/pkg/auth/roles"

	"github.com/go-playground/assert/v2"
)

const (
	tenant      = "11111111-1111-1111-1111-111111111111"
	terminalApp = "bbbbbbbb-0000-0000-0000-000000000002"
)

var token = auth.Token{Issuer: "https://login.microsoftonline.com/" + tenant + "/v2.0", Subject: "pairwise", Expiry: time.Unix(2000000000, 0)}

func testConfig() Config {
	return Config{TenantIds: []string{tenant}, ApiClientId: "api", RequiredUserScope: "access_as_user"}.withDefaults()
}

func appClaims() claims {
	return claims{TenantId: tenant, ObjectId: "sp-oid", AuthorizedParty: terminalApp, IdentityType: "app", Roles: []string{"Playbook.Execute"}}
}

func userClaims() claims {
	return claims{TenantId: tenant, ObjectId: "user-oid", AuthorizedParty: "gui", Scope: "openid access_as_user",
		Roles: []string{"Playbook.Read"}, Name: "Alice", PreferredUsername: "alice@example.com"}
}

func TestAppToken(t *testing.T) {
	principal, err := testConfig().principal(tenant, token, appClaims())

	assert.Equal(t, err, nil)
	assert.Equal(t, principal.Kind, auth.KindApp)
	assert.Equal(t, principal.ClientId, terminalApp)
	assert.Equal(t, principal.Subject, "sp-oid")
	assert.Equal(t, principal.Self(), "entra:"+tenant+":app:"+terminalApp)
	assert.Equal(t, principal.Permissions, map[roles.Permission]bool{roles.PlaybookRead: true, roles.PlaybookExecute: true})
	assert.Equal(t, principal.ExpiresAt, token.Expiry)
}

func TestUserToken(t *testing.T) {
	principal, err := testConfig().principal(tenant, token, userClaims())

	assert.Equal(t, err, nil)
	assert.Equal(t, principal.Kind, auth.KindUser)
	assert.Equal(t, principal.Name, "Alice")
	assert.Equal(t, principal.Email, "alice@example.com")
	assert.Equal(t, principal.Self(), "entra:"+tenant+":user:user-oid")
	assert.Equal(t, principal.Has(roles.PlaybookRead), true)
	assert.Equal(t, principal.Has(roles.PlaybookExecute), false)
}

func TestKindWithoutIdtypClaim(t *testing.T) {
	app := appClaims()
	app.IdentityType = ""
	user := userClaims()

	assert.Equal(t, app.kind(), auth.KindApp)
	assert.Equal(t, user.kind(), auth.KindUser)
}

func TestV1ClientIdComesFromAppid(t *testing.T) {
	v1 := appClaims()
	v1.AuthorizedParty = ""
	v1.AppId = terminalApp

	principal, err := testConfig().principal(tenant, token, v1)

	assert.Equal(t, err, nil)
	assert.Equal(t, principal.ClientId, terminalApp)
}

func TestTenantMustMatchIssuer(t *testing.T) {
	foreign := appClaims()
	foreign.TenantId = "22222222-2222-2222-2222-222222222222"

	_, err := testConfig().principal(tenant, token, foreign)

	assert.NotEqual(t, err, nil)
}

func TestAllowedClientIds(t *testing.T) {
	config := testConfig()
	config.AllowedClientIds = []string{terminalApp}
	other := appClaims()
	other.AuthorizedParty = "someone-else"

	_, err := config.principal(tenant, token, appClaims())
	assert.Equal(t, err, nil)
	_, err = config.principal(tenant, token, other)
	assert.NotEqual(t, err, nil)
}

func TestUserTokenNeedsRequiredScope(t *testing.T) {
	wrongScope := userClaims()
	wrongScope.Scope = "User.Read"

	_, err := testConfig().principal(tenant, token, wrongScope)
	assert.NotEqual(t, err, nil)

	config := testConfig()
	config.RequiredUserScope = ""
	_, err = config.principal(tenant, token, wrongScope)
	assert.Equal(t, err, nil)
}

func TestGroupsBecomeSubjects(t *testing.T) {
	user := userClaims()
	user.Groups = []string{"G-Blue", "g-red"}

	principal, err := testConfig().principal(tenant, token, user)

	assert.Equal(t, err, nil)
	assert.Equal(t, principal.Subjects, []string{
		"entra:" + tenant + ":user:user-oid",
		"entra:" + tenant + ":group:g-blue",
		"entra:" + tenant + ":group:g-red",
	})
	assert.Equal(t, principal.GroupsIncomplete, false)
}

func TestGroupOverflowIsFlagged(t *testing.T) {
	user := userClaims()
	user.ClaimNames = map[string]string{"groups": "src1"}

	principal, err := testConfig().principal(tenant, token, user)

	assert.Equal(t, err, nil)
	assert.Equal(t, principal.GroupsIncomplete, true)
}

func TestNoRolesNoPermissions(t *testing.T) {
	app := appClaims()
	app.Roles = nil

	principal, err := testConfig().principal(tenant, token, app)

	assert.Equal(t, err, nil)
	assert.Equal(t, len(principal.Permissions), 0)
}
