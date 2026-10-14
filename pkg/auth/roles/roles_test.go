package roles

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

var testRoles = RoleMap{
	"reader": {PlaybookRead},
	"author": {PlaybookRead, PlaybookWrite},
	"runner": {PlaybookRead, PlaybookExecute},
	"admin":  All(),
}

func TestPermissionsAreTheUnionOfRoles(t *testing.T) {
	granted := testRoles.Permissions([]string{"author", "runner"})

	assert.Equal(t, granted, map[Permission]bool{PlaybookRead: true, PlaybookWrite: true, PlaybookExecute: true})
}

func TestUnknownRolesGrantNothing(t *testing.T) {
	granted := testRoles.Permissions([]string{"Directory.ReadWrite.All", "Reader"})

	assert.Equal(t, len(granted), 0)
}

func TestEmptyRoleMapGrantsNothing(t *testing.T) {
	granted := RoleMap{}.Permissions([]string{"admin"})

	assert.Equal(t, len(granted), 0)
}

func TestAllListsEveryPermission(t *testing.T) {
	granted := testRoles.Permissions([]string{"admin"})

	assert.Equal(t, len(granted), 4)
	for _, permission := range All() {
		assert.Equal(t, granted[permission], true)
	}
}
