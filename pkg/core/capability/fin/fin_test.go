package fin

import (
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	model "soarca/pkg/models/fin"
	"soarca/test/unittest/mocks/mock_finprotocol"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestFinExecution(t *testing.T) {
	mockFinProtocol := new(mock_finprotocol.MockFinProtocol)
	//mockGuid := new(mock_guid.Mock_Guid)
	finCapability := New(mockFinProtocol)

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")

	var metadata = execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}

	command := cacao.Command{Type: "soarca-fin", Command: "test command"}
	auth := cacao.AuthenticationInformation{}
	auth.Name = "some auth"
	auth.Username = "user"
	auth.Password = "password"
	target := cacao.AgentTarget{}
	variable1 := cacao.Variable{Type: "int", Name: "output", Value: "10"}

	//var id, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	expectedCommand := model.Command{}
	expectedCommand.Type = "command"
	expectedCommand.CommandSubstructure.Context.ExecutionId = executionId.String()

	expectedCommand.CommandSubstructure.Authentication = auth
	expectedCommand.CommandSubstructure.Command = "test command"
	expectedCommand.CommandSubstructure.Context.Timeout = 1
	expectedCommand.CommandSubstructure.Context.PlaybookId = playbookId.String()
	expectedCommand.CommandSubstructure.Context.StepId = stepId.String()

	variable2 := cacao.Variable{Type: "string", Name: "input", Value: "some value"}
	inputVariable := map[string]cacao.Variable{"input_variable": variable2}
	expectedCommand.CommandSubstructure.Variables = inputVariable

	//expectedCommand.CommandSubstructure.Context.GeneratedOn = ""

	expectedVariableMap := cacao.NewVariables(variable1)

	data := capability.Context{
		Command:        command,
		Authentication: auth,
		Target:         target,
		Variables:      inputVariable,
	}

	//mockGuid.On("New").Return(id)
	mockFinProtocol.On("SendCommand", expectedCommand).Return(expectedVariableMap, nil)
	result, err := finCapability.Execute(metadata, data)

	assert.Equal(t, err, nil)
	assert.Equal(t, result, expectedVariableMap)

}
