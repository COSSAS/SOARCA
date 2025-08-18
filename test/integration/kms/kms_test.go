package kms

import (
	"fmt"
	"os"
	"path"
	"soarca/internal/database/memory"
	"soarca/pkg/core/capability"
	"soarca/pkg/core/capability/ssh"
	"soarca/pkg/keymanagement"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var globalKeyManagement *keymanagement.KeyManagement

func init() {
	globalKeyManagement = keymanagement.InitKeyManagement(memory.NewKeyManagementDatabase())
}

const testkey string = "test"

func testkey_dir() string {
	return path.Join("..", "..", "..", "deployments", "docker", "testing", "ssh-kms-test")
}
func addTestKey(t *testing.T) {
	pubkey_path := path.Join(testkey_dir(), testkey+".pub")
	privkey_path := path.Join(testkey_dir(), testkey)
	pubkey, err := os.ReadFile(pubkey_path)
	assert.Nil(t, err)
	privkey, err := os.ReadFile(privkey_path)
	assert.Nil(t, err)
	assert.Nil(t, globalKeyManagement.Insert(string(pubkey), string(privkey), "", testkey))
}

func TestSshConnection(t *testing.T) {
	sshCapability := ssh.SshCapability{Keys: globalKeyManagement}
	addTestKey(t)

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ls -la",
	}

	expectedAuthenticationInformation := cacao.AuthenticationInformation{
		ID:               "some-authid-1",
		Type:             "user-auth",
		Username:         "sshtest",
		Kms:              true,
		KmsKeyIdentifier: testkey,
	}

	expectedTarget := cacao.AgentTarget{
		Type:               "ssh",
		Address:            map[cacao.NetAddressType][]string{"ipv4": {"localhost"}},
		Port:               "2223",
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
