package powershell_integration_test

import (
	"fmt"
	"soarca/internal/capability/powershell"
	"soarca/models/cacao"
	"soarca/models/execution"
	"testing"

	"github.com/google/uuid"
)

func TestPowershellConnection(t *testing.T) {
	capability := powershell.New()

	expectedCommand := cacao.Command{
		Type:    "powershell",
		Command: "Get-Acl",
	}

	expectedAuthenticationInformation := cacao.AuthenticationInformation{
		ID:       "some-authid-1",
		Type:     "user-auth",
		Username: "User",
		Password: "Password899!"}

	expectedTarget := cacao.AgentTarget{
		Type:               "ssh",
		Address:            map[cacao.NetAddressType][]string{"ipv4": {"127.0.0.1"}},
		Port:               "5985",
		AuthInfoIdentifier: "some-authid-1",
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId = "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	var stepId = "step--81eff59f-d084-4324-9e0a-59e353dbd28f"
	var metadata = execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId, StepId: stepId}
	results, err := capability.Execute(metadata,
		expectedCommand,
		expectedAuthenticationInformation,
		expectedTarget,
		cacao.NewVariables())
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	fmt.Println(results)

}
