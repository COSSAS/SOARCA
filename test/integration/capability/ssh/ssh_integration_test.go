package ssh_integration_test

import (
	"fmt"
	"soarca/pkg/core/capability"
	"soarca/pkg/core/capability/ssh"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
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
		Address: map[cacao.NetAddressType][]string{"ipv4": {"localhost"}},
		// Port:               "22",
		AuthInfoIdentifier: "some-authid-1",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId = "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	var stepId = "step--81eff59f-d084-4324-9e0a-59e353dbd28f"
	var metadata = execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId, StepId: stepId}
	data := capability.Context{
		Command:        expectedCommand,
		Target:         expectedTarget,
		Authentication: expectedAuthenticationInformation,
		Variables:      cacao.NewVariables(expectedVariables),
	}
	results, err := sshCapability.Execute(metadata,
		data)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	fmt.Println(results)

}

func TestSshConnectionToNonExistingServer(t *testing.T) {
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
		Address: map[cacao.NetAddressType][]string{"ipv4": {"10.10.10.10"}},
		// Port:               "22",
		AuthInfoIdentifier: "some-authid-1",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId = "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	var stepId = "step--81eff59f-d084-4324-9e0a-59e353dbd28f"
	var metadata = execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId, StepId: stepId}
	data := capability.Context{
		Command:        expectedCommand,
		Target:         expectedTarget,
		Authentication: expectedAuthenticationInformation,
		Variables:      cacao.NewVariables(expectedVariables),
	}
	results, err := sshCapability.Execute(metadata,
		data)
	assert.NotEqual(t, err, nil)

	fmt.Println(results)

}
