package ssh

import (
	"errors"
	"soarca/pkg/models/cacao"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestStripSshPrependWithPrepend(t *testing.T) {
	command := "ssh ls -la"
	result := StripSshPrepend(command)
	assert.Equal(t, result, "ls -la")
}

func TestStripSshPrependWithoutPrepend(t *testing.T) {
	command := "ls -la"
	result := StripSshPrepend(command)
	assert.Equal(t, result, "ls -la")
}

func TestAuthenticationValidationUserAuth(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "user-auth", Username: "root", Password: "password"}
	result := CheckSshAuthenticationInfo(auth)
	assert.Equal(t, result, nil)
}

func TestAuthenticationValidationUserAuthMissingPassword(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "user-auth", Username: "root"}
	result := CheckSshAuthenticationInfo(auth)
	err := errors.New("password is empty")
	assert.Equal(t, result, err)
}

func TestAuthenticationValidationUserAuthSpacesAsPassword(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "user-auth", Username: "root", Password: "   "}
	result := CheckSshAuthenticationInfo(auth)
	err := errors.New("password is empty")
	assert.Equal(t, result, err)
}

func TestAuthenticationValidationUserAuthSpacesAsUser(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "user-auth", Username: "   ", Password: "password"}
	result := CheckSshAuthenticationInfo(auth)
	err := errors.New("username is empty")
	assert.Equal(t, result, err)
}

func TestAuthenticationValidationPrivateKeyAuth(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "private-key", Username: "root", PrivateKey: "someprivatekey"}
	result := CheckSshAuthenticationInfo(auth)
	assert.Equal(t, result, nil)
}

func TestAuthenticationValidationPrivateKeyAuthMissingKey(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "private-key", Username: "root"}
	result := CheckSshAuthenticationInfo(auth)
	err := errors.New("private key is not set")
	assert.Equal(t, result, err)
}

func TestAuthenticationValidationPrivateKeyAuthSpacesAsKey(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "private-key", Username: "root", PrivateKey: "   "}
	result := CheckSshAuthenticationInfo(auth)
	err := errors.New("private key is not set")
	assert.Equal(t, result, err)
}

func TestAuthenticationValidationPrivateKeyAuthSpacesAsUser(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "private-key", Username: "   ", PrivateKey: "someprivatekey"}
	result := CheckSshAuthenticationInfo(auth)
	err := errors.New("username is empty")
	assert.Equal(t, result, err)
}

func TestAddressAndPortCombination(t *testing.T) {
	ipv4 := map[cacao.NetAddressType][]string{"ipv4": {"134.221.49.62"}}
	port := "22"
	expectedFqdn := "134.221.49.62:22"
	result := CombinePortAndAddress(ipv4, port)
	assert.Equal(t, result, expectedFqdn)
}
func TestAddressAndPortCombinationNoPort(t *testing.T) {
	ipv4 := map[cacao.NetAddressType][]string{"ipv4": {"134.221.49.62"}}
	port := ""
	expectedFqdn := "134.221.49.62:22"
	result := CombinePortAndAddress(ipv4, port)
	assert.Equal(t, result, expectedFqdn)
}

func TestAddressAndPortCombinationNoAddress(t *testing.T) {
	ipv4 := map[cacao.NetAddressType][]string{}
	port := "22"
	expectedFqdn := ""
	result := CombinePortAndAddress(ipv4, port)
	assert.Equal(t, result, expectedFqdn)
}
func TestAddressAndPortCombinationNoIpv4Address(t *testing.T) {
	ipv4 := map[cacao.NetAddressType][]string{"invalid": {"feed::0001"}}
	port := "22"
	expectedFqdn := ""
	result := CombinePortAndAddress(ipv4, port)
	assert.Equal(t, result, expectedFqdn)
}
