package action_executor_test

import (
	"errors"
	"testing"

	"soarca/internal/capability"
	"soarca/internal/executors/action"
	"soarca/models/cacao"
	"soarca/models/execution"
	"soarca/test/unittest/mocks/mock_capability"
	"soarca/test/unittest/mocks/mock_reporter"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestExecuteStep(t *testing.T) {
	mock_ssh := new(mock_capability.Mock_Capability)
	mock_http := new(mock_capability.Mock_Capability)
	mock_reporter := new(mock_reporter.Mock_Reporter)

	capabilities := map[string]capability.ICapability{"mock-ssh": mock_ssh, "http-api": mock_http}

	executerObject := action.New(capabilities, mock_reporter)
	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId := "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	stepId := "step--81eff59f-d084-4324-9e0a-59e353dbd28f"

	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId, StepId: stepId}

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
		ID:   "1",
	}

	expectedTarget := cacao.AgentTarget{
		ID:                 "target1",
		Name:               "sometarget",
		AuthInfoIdentifier: "1",
	}

	agent := cacao.AgentTarget{
		Type: "ssh",
		Name: "mock-ssh",
	}

	step := cacao.Step{
		Type:          cacao.StepTypeAction,
		Name:          "action test",
		ID:            stepId,
		Description:   "",
		StepVariables: cacao.NewVariables(expectedVariables),
		Commands:      []cacao.Command{expectedCommand},
		Agent:         "mock-ssh",
		Targets:       []string{"target1"},
	}

	actionMetadata := action.PlaybookStepMetadata{
		Step:      step,
		Targets:   map[string]cacao.AgentTarget{expectedTarget.ID: expectedTarget},
		Auth:      map[string]cacao.AuthenticationInformation{expectedAuth.ID: expectedAuth},
		Agent:     agent,
		Variables: cacao.NewVariables(expectedVariables),
	}
	mock_reporter.On("ReportStepStart", executionId, step, cacao.NewVariables(expectedVariables)).Return()

	mock_reporter.On("ReportStepEnd", executionId, step, cacao.NewVariables(expectedVariables), nil).Return()
	mock_ssh.On("Execute",
		metadata,
		expectedCommand,
		expectedAuth,
		expectedTarget,
		cacao.NewVariables(expectedVariables)).
		Return(cacao.NewVariables(expectedVariables),
			nil)

	_, err := executerObject.Execute(metadata,
		actionMetadata)

	assert.Equal(t, err, nil)
	mock_reporter.AssertExpectations(t)
	mock_ssh.AssertExpectations(t)
}

func TestExecuteActionStep(t *testing.T) {
	mock_ssh := new(mock_capability.Mock_Capability)
	mock_http := new(mock_capability.Mock_Capability)
	mock_reporter := new(mock_reporter.Mock_Reporter)

	capabilities := map[string]capability.ICapability{"ssh": mock_ssh, "http-api": mock_http}

	executerObject := action.New(capabilities, mock_reporter)
	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId := "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	stepId := "step--81eff59f-d084-4324-9e0a-59e353dbd28f"

	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId, StepId: stepId}

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

	_, err := executerObject.ExecuteActionStep(metadata,
		expectedCommand,
		expectedAuth,
		expectedTarget,
		cacao.NewVariables(expectedVariables),
		agent)

	assert.Equal(t, err, nil)
	mock_reporter.AssertExpectations(t)
	mock_ssh.AssertExpectations(t)
}

func TestNonExistingCapabilityStep(t *testing.T) {
	mock_ssh := new(mock_capability.Mock_Capability)
	mock_http := new(mock_capability.Mock_Capability)

	capabilities := map[string]capability.ICapability{"ssh": mock_ssh, "http-api": mock_http}

	executerObject := action.New(capabilities, new(mock_reporter.Mock_Reporter))
	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId := "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	stepId := "step--81eff59f-d084-4324-9e0a-59e353dbd28f"

	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId, StepId: stepId}

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

	_, err := executerObject.ExecuteActionStep(metadata,
		expectedCommand,
		expectedAuth,
		expectedTarget,
		cacao.NewVariables(expectedVariables),
		agent)

	assert.Equal(t, err, errors.New("capability: non-existing is not available in soarca"))
	mock_ssh.AssertExpectations(t)
}

func TestVariableInterpolation(t *testing.T) {
	mock_capability1 := new(mock_capability.Mock_Capability)

	capabilities := map[string]capability.ICapability{"cap1": mock_capability1}

	executerObject := action.New(capabilities, new(mock_reporter.Mock_Reporter))
	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId := "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	stepId := "step--81eff59f-d084-4324-9e0a-59e353dbd28f"

	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId, StepId: stepId}

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

	varHttpContent := cacao.Variable{
		Type:  cacao.VariableTypeString,
		Name:  "__http-api-content__",
		Value: "some content of the body",
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

	varheader1 := cacao.Variable{
		Type:  "string",
		Name:  "__header_var1__",
		Value: "headerinfo one",
	}

	varheader2 := cacao.Variable{
		Type:  "string",
		Name:  "__header_var2__",
		Value: "headerinfo two",
	}

	inputHeaders := cacao.Headers{
		"header1": []string{"__header_var1__:value", "__header_var2__:value"},
	}

	expectedHeaders := cacao.Headers{
		"header1": []string{"headerinfo one", "headerinfo two"},
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
		cacao.NewVariables(var1, var2, var3, varUser, varPassword, varOauth, varPrivateKey, varToken, varUserId, varheader1, varheader2)).
		Return(cacao.NewVariables(var1),
			nil)

	_, err := executerObject.ExecuteActionStep(metadata,
		inputCommand,
		inputAuth,
		inputTarget,
		cacao.NewVariables(var1, var2, var3, varUser, varPassword, varOauth, varPrivateKey, varToken, varUserId, varheader1, varheader2),
		agent)

	assert.Equal(t, err, nil)
	mock_capability1.AssertExpectations(t)
	assert.Equal(t, inputCommand.Command, "ssh __var1__:value")

	httpCommand := cacao.Command{
		Type:       "http-api",
		Command:    "GET / HTTP1.1",
		Content:    "__http-api-content__:value",
		ContentB64: "__http-api-content__:value",
		Headers:    inputHeaders,
	}

	expectedHttpCommand := cacao.Command{
		Type:       "http-api",
		Command:    "GET / HTTP1.1",
		Content:    "some content of the body",
		ContentB64: "some content of the body",
		Headers:    expectedHeaders,
	}

	metadataHttp := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId, StepId: stepId}

	mock_capability1.On("Execute",
		metadataHttp,
		expectedHttpCommand,
		expectedAuth,
		expectedTarget,
		cacao.NewVariables(varHttpContent, varheader1, varheader2)).
		Return(cacao.NewVariables(var1),
			nil)

	_, err = executerObject.ExecuteActionStep(metadata,
		httpCommand,
		expectedAuth,
		expectedTarget,
		cacao.NewVariables(varHttpContent, varheader1, varheader2),
		agent)

	assert.Equal(t, err, nil)
	mock_capability1.AssertExpectations(t)

}
