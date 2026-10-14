package auth

import (
	"testing"

	"soarca/pkg/auth/roles"

	"github.com/go-playground/assert/v2"
)

func TestHas(t *testing.T) {
	principal := &Principal{Permissions: map[roles.Permission]bool{roles.PlaybookRead: true}}

	assert.Equal(t, principal.Has(roles.PlaybookRead), true)
	assert.Equal(t, principal.Has(roles.PlaybookExecute), false)
}

func TestNilPrincipalHasNothing(t *testing.T) {
	var principal *Principal

	assert.Equal(t, principal.Has(roles.Admin), false)
	assert.Equal(t, principal.Self(), "")
}

func TestSubjects(t *testing.T) {
	principal := &Principal{Subjects: []string{
		UserSubject("entra:tenant", "oid-1"),
		GroupSubject("entra:tenant", "group-1"),
	}}

	assert.Equal(t, principal.Self(), "entra:tenant:user:oid-1")
	assert.Equal(t, principal.Subjects[1], "entra:tenant:group:group-1")
	assert.Equal(t, AppSubject("apikey", "webhook"), "apikey:app:webhook")
}
