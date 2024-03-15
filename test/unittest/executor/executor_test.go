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

func TestVariableInterpolation(t *testing.T) {

	mock_capability1 := new(mock_capability.Mock_Capability)

	capabilities := map[string]capability.ICapability{"cap1": mock_capability1}

	executerObject := executer.New(capabilities)
	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId = "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	var stepId = "step--81eff59f-d084-4324-9e0a-59e353dbd28f"

	var metadata = execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId, StepId: stepId}

	inputCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh __var1__:value",
	}

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	var1 := cacao.Variable{
		Type:  "string",
		Name:  "__var1__",
		Value: "ls -la",
	}

	var2 := cacao.Variable{
		Type:  "string",
		Name:  "__var2__",
		Value: "https://httpbin.org/put",
	}

	var3 := cacao.Variable{
		Type:  "string",
		Name:  "__var3__",
		Value: "1.3.3.7",
	}

	varUser := cacao.Variable{
		Type:  "string",
		Name:  "__user__",
		Value: "soarca-user",
	}
	varPassword := cacao.Variable{
		Type:  "string",
		Name:  "__password__",
		Value: "soarca-password",
	}
	varToken := cacao.Variable{
		Type:  "string",
		Name:  "__token__",
		Value: "soarca-token",
	}
	varUserId := cacao.Variable{
		Type:  "string",
		Name:  "__userid__",
		Value: "soarca-userid",
	}

	varOauth := cacao.Variable{
		Type:  "string",
		Name:  "__oauth__",
		Value: "soarca-oauth",
	}
	varPrivateKey := cacao.Variable{
		Type:  "string",
		Name:  "__privatekey__",
		Value: "soarca-privatekey",
	}

	inputAuth := cacao.AuthenticationInformation{
		Name:        "soarca",
		Username:    "__user__:value",
		UserId:      "__userid__:value",
		Password:    "__password__:value",
		PrivateKey:  "__privatekey__:value",
		Token:       "__token__:value",
		OauthHeader: "__oauth__:value",
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name:        "soarca",
		Username:    "soarca-user",
		UserId:      "soarca-userid",
		Password:    "soarca-password",
		PrivateKey:  "soarca-privatekey",
		Token:       "soarca-token",
		OauthHeader: "soarca-oauth",
	}

	inputTarget := cacao.AgentTarget{
		Name: "sometarget",
		Address: map[cacao.NetAddressType][]string{
			cacao.Url:  {"__var2__:value"},
			cacao.IPv4: {"__var3__:value"},
		},
	}

	expectedTarget := cacao.AgentTarget{
		Name: "sometarget",
		Address: map[cacao.NetAddressType][]string{
			cacao.Url:  {"https://httpbin.org/put"},
			cacao.IPv4: {"1.3.3.7"},
		},
	}

	agent := cacao.AgentTarget{
		Type: "ssh",
		Name: "cap1",
	}

	mock_capability1.On("Execute",
		metadata,
		expectedCommand,
		expectedAuth,
		expectedTarget,
		cacao.NewVariables(var1, var2, var3, varUser, varPassword, varOauth, varPrivateKey, varToken, varUserId)).
		Return(cacao.NewVariables(var1),
			nil)

	_, _, err := executerObject.Execute(metadata,
		inputCommand,
		inputAuth,
		inputTarget,
		cacao.NewVariables(var1, var2, var3, varUser, varPassword, varOauth, varPrivateKey, varToken, varUserId),
		agent)

	assert.Equal(t, err, nil)
	mock_capability1.AssertExpectations(t)
	assert.Equal(t, inputCommand.Command, "ssh __var1__:value")
}
