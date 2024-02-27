package ssh_test

import (
	"errors"
	"fmt"
	"soarca/internal/capability/ssh"
	"soarca/models/cacao"
	"soarca/models/execution"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestSshConnection(t *testing.T) {
	sshCapability := new(ssh.SshCapability)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ls -la",
	}

	expectedAuthenticationInformation := cacao.AuthenticationInformation{
		ID:       "some-authid-1",
		Type:     "user-auth",
		Username: "sshtest",
		Password: "pdKY77qNxpI5MAizirtjCVOcm0KFKIs"}

	expectedTarget := cacao.AgentTarget{
		Type:    "ssh",
		Address: map[string][]string{"ipv4": {"localhost"}},
		// Port:               "22",
		AuthInfoIdentifier: "some-authid-1",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}
	results, err := sshCapability.Execute(metadata,
		expectedCommand,
		expectedAuthenticationInformation,
		expectedTarget,
		map[string]cacao.Variable{expectedVariables.Name: expectedVariables})
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	fmt.Println(results)

}

func TestStripSshPrependWithPrepend(t *testing.T) {
	command := "ssh ls -la"
	result := ssh.StripSshPrepend(command)
	assert.Equal(t, result, "ls -la")
}

func TestStripSshPrependWithoutPrepend(t *testing.T) {
	command := "ls -la"
	result := ssh.StripSshPrepend(command)
	assert.Equal(t, result, "ls -la")
}

func TestAuthenticationValidationUserAuth(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "user-auth", Username: "root", Password: "password"}
	result := ssh.CheckSshAuthenticationInfo(auth)
	assert.Equal(t, result, nil)
}

func TestAuthenticationValidationUserAuthMissingPassword(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "user-auth", Username: "root"}
	result := ssh.CheckSshAuthenticationInfo(auth)
	err := errors.New("password is empty")
	assert.Equal(t, result, err)
}

func TestAuthenticationValidationUserAuthSpacesAsPassword(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "user-auth", Username: "root", Password: "   "}
	result := ssh.CheckSshAuthenticationInfo(auth)
	err := errors.New("password is empty")
	assert.Equal(t, result, err)
}

func TestAuthenticationValidationUserAuthSpacesAsUser(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "user-auth", Username: "   ", Password: "password"}
	result := ssh.CheckSshAuthenticationInfo(auth)
	err := errors.New("username is empty")
	assert.Equal(t, result, err)
}

func TestAuthenticationValidationPrivateKeyAuth(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "private-key", Username: "root", PrivateKey: "someprivatekey"}
	result := ssh.CheckSshAuthenticationInfo(auth)
	assert.Equal(t, result, nil)
}

func TestAuthenticationValidationPrivateKeyAuthMissingKey(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "private-key", Username: "root"}
	result := ssh.CheckSshAuthenticationInfo(auth)
	err := errors.New("private key is not set")
	assert.Equal(t, result, err)
}

func TestAuthenticationValidationPrivateKeyAuthSpacesAsKey(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "private-key", Username: "root", PrivateKey: "   "}
	result := ssh.CheckSshAuthenticationInfo(auth)
	err := errors.New("private key is not set")
	assert.Equal(t, result, err)
}

func TestAuthenticationValidationPrivateKeyAuthSpacesAsUser(t *testing.T) {
	auth := cacao.AuthenticationInformation{Type: "private-key", Username: "   ", PrivateKey: "someprivatekey"}
	result := ssh.CheckSshAuthenticationInfo(auth)
	err := errors.New("username is empty")
	assert.Equal(t, result, err)
}

func TestAddressAndPortCombination(t *testing.T) {
	ipv4 := map[string][]string{"ipv4": {"134.221.49.62"}}
	port := "22"
	expectedFqdn := "134.221.49.62:22"
	result := ssh.CombinePortAndAddress(ipv4, port)
	assert.Equal(t, result, expectedFqdn)
}
func TestAddressAndPortCombinationNoPort(t *testing.T) {
	ipv4 := map[string][]string{"ipv4": {"134.221.49.62"}}
	port := ""
	expectedFqdn := "134.221.49.62:22"
	result := ssh.CombinePortAndAddress(ipv4, port)
	assert.Equal(t, result, expectedFqdn)
}

func TestAddressAndPortCombinationNoAddress(t *testing.T) {
	ipv4 := map[string][]string{}
	port := "22"
	expectedFqdn := ""
	result := ssh.CombinePortAndAddress(ipv4, port)
	assert.Equal(t, result, expectedFqdn)
}
func TestAddressAndPortCombinationNoIpv4Address(t *testing.T) {
	ipv4 := map[string][]string{"invallid": {"feed::0001"}}
	port := "22"
	expectedFqdn := ""
	result := ssh.CombinePortAndAddress(ipv4, port)
	assert.Equal(t, result, expectedFqdn)
}
