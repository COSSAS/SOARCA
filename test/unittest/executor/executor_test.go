package executor_test

import (
	"errors"
	"soarca/internal/capability"
	"soarca/internal/executer"
	"soarca/models/cacao"
	"soarca/models/execution"
	"soarca/test/unittest/mocks/mock_capability"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestExecuteStep(t *testing.T) {

	mock_ssh := new(mock_capability.Mock_Capability)
	mock_http := new(mock_capability.Mock_Capability)

	capabilities := map[string]capability.ICapability{"ssh": mock_ssh, "http-api": mock_http}

	executerObject := executer.New(capabilities)
	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId = "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	var stepId = "step--81eff59f-d084-4324-9e0a-59e353dbd28f"

	var metadata = execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId, StepId: stepId}

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name: "user",
	}

	expectedTarget := cacao.AgentTarget{
		Name: "sometarget",
	}

	agent := cacao.AgentTarget{
		Type: "ssh",
		Name: "ssh",
	}

	mock_ssh.On("Execute",
		metadata,
		expectedCommand,
		expectedAuth,
		expectedTarget,
		cacao.NewVariables(expectedVariables)).
		Return(cacao.NewVariables(expectedVariables),
			nil)

	_, _, err := executerObject.Execute(metadata,
		expectedCommand,
		expectedAuth,
		expectedTarget,
		cacao.NewVariables(expectedVariables),
		agent)

	assert.Equal(t, err, nil)
	mock_ssh.AssertExpectations(t)
}

func TestNonExistingCapabilityStep(t *testing.T) {

	mock_ssh := new(mock_capability.Mock_Capability)
	mock_http := new(mock_capability.Mock_Capability)

	capabilities := map[string]capability.ICapability{"ssh": mock_ssh, "http-api": mock_http}

	executerObject := executer.New(capabilities)
	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId = "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	var stepId = "step--81eff59f-d084-4324-9e0a-59e353dbd28f"

	var metadata = execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId, StepId: stepId}

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name: "user",
	}

	expectedTarget := cacao.AgentTarget{
		Name: "sometarget",
	}

	agent := cacao.AgentTarget{
		Type: "ssh",
		Name: "non-existing",
	}

	_, _, err := executerObject.Execute(metadata,
		expectedCommand,
		expectedAuth,
		expectedTarget,
		cacao.NewVariables(expectedVariables),
		agent)

	assert.Equal(t, err, errors.New("executor is not available in soarca"))
	mock_ssh.AssertExpectations(t)
}
