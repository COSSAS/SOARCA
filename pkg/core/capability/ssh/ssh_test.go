package ssh

import (
	"errors"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
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

func TestSshExecuteNoTargetsSkipsWithoutError(t *testing.T) {
	sshCapability := &SshCapability{}

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId, _ := uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	stepId, _ := uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")

	command := cacao.Command{Type: "ssh", Command: "ls -la"}

	metadata := execution.Metadata{
		ExecutionId: executionId,
		PlaybookId:  playbookId.String(),
		StepId:      stepId.String(),
	}

	data := capability.Context{
		Commands: []cacao.Command{command},
		Targets:  []capability.ResolvedTarget{},
	}

	results, err := sshCapability.Execute(metadata, data)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, len(results), 0)
}
