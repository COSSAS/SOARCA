package mqtt_test

import (
	"fmt"
	"soarca/internal/fin/protocol"
	model "soarca/pkg/models/fin"
	"soarca/pkg/utils/guid"
	"testing"

	"github.com/google/uuid"
)

func TestConnect(t *testing.T) {
	// used for manual testing

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId = "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	var stepId = "step--81eff59f-d084-4324-9e0a-59e353dbd28f"
	guid := new(guid.Guid)
	prot := protocol.FinProtocol{Guid: guid, Topic: protocol.Topic("testing"), Broker: "localhost", Port: 1883}
	expectedCommand := model.NewCommand()
	expectedCommand.CommandSubstructure.Context.Timeout = 10
	expectedCommand.CommandSubstructure.Context.ExecutionId = executionId.String()

	expectedCommand.CommandSubstructure.Command = "test command"
	expectedCommand.CommandSubstructure.Context.PlaybookId = playbookId
	expectedCommand.CommandSubstructure.Context.StepId = stepId

	result, err := prot.SendCommand(expectedCommand)
	if err != nil {
		t.Fail()
	}
	fmt.Println(result)
	fmt.Println(err)

}
