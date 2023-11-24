package executor_test

import (
	"errors"
	"soarca/internal/capability"
	"soarca/internal/executer"
	"soarca/models/cacao"
	"soarca/test/mocks/mock_capability"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestExecuteStep(t *testing.T) {

	mock_ssh := new(mock_capability.Mock_Capability)
	mock_http := new(mock_capability.Mock_Capability)

	capabilities := map[string]capability.ICapability{"ssh": mock_ssh, "http-api": mock_http}

	executerObject := executer.New(capabilities)
	var id, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variables{
		ObjectType: "string",
		Name:       "var1",
		Value:      "testing",
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name: "user",
	}

	expectedTarget := cacao.Target{
		Name: "sometarget",
	}

	agent := cacao.Target{
		Type: "ssh",
		Name: "ssh",
	}

	mock_ssh.On("Execute",
		id,
		expectedCommand,
		expectedAuth,
		expectedTarget,
		map[string]cacao.Variables{expectedVariables.Name: expectedVariables}).
		Return(map[string]cacao.Variables{expectedVariables.Name: expectedVariables},
			nil)

	_, _, err := executerObject.Execute(id,
		expectedCommand,
		expectedAuth,
		expectedTarget,
		map[string]cacao.Variables{expectedVariables.Name: expectedVariables},
		agent)

	assert.Equal(t, err, nil)
	mock_ssh.AssertExpectations(t)
}

func TestNonExistingCapabilityStep(t *testing.T) {

	mock_ssh := new(mock_capability.Mock_Capability)
	mock_http := new(mock_capability.Mock_Capability)

	capabilities := map[string]capability.ICapability{"ssh": mock_ssh, "http-api": mock_http}

	executerObject := executer.New(capabilities)
	var id, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variables{
		ObjectType: "string",
		Name:       "var1",
		Value:      "testing",
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name: "user",
	}

	expectedTarget := cacao.Target{
		Name: "sometarget",
	}

	agent := cacao.Target{
		Type: "ssh",
		Name: "non-existing",
	}

	_, _, err := executerObject.Execute(id,
		expectedCommand,
		expectedAuth,
		expectedTarget,
		map[string]cacao.Variables{expectedVariables.Name: expectedVariables},
		agent)

	assert.Equal(t, err, errors.New("executor is not available in soarca"))
	mock_ssh.AssertExpectations(t)
}
