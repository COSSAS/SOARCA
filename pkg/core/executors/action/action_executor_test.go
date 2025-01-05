package action

import (
	"errors"
	"testing"
	"time"

	"soarca/pkg/core/capability"
	"soarca/pkg/core/executors"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/test/unittest/mocks/mock_capability"
	"soarca/test/unittest/mocks/mock_reporter"
	mock_time "soarca/test/unittest/mocks/mock_utils/time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestExecuteStep(t *testing.T) {
	mock_ssh := new(mock_capability.Mock_Capability)
	mock_http := new(mock_capability.Mock_Capability)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

	capabilities := map[string]capability.ICapability{"mock-ssh": mock_ssh, "http-api": mock_http}

	executorObject := New(capabilities, mock_reporter, mock_time)
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

	actionMetadata := executors.PlaybookStepMetadata{
		Step:      step,
		Targets:   map[string]cacao.AgentTarget{expectedTarget.ID: expectedTarget},
		Auth:      map[string]cacao.AuthenticationInformation{expectedAuth.ID: expectedAuth},
		Agent:     agent,
		Variables: cacao.NewVariables(expectedVariables),
	}

	context1 := capability.Context{
		Command:        expectedCommand,
		Authentication: expectedAuth,
		Target:         expectedTarget,
		Variables:      cacao.NewVariables(expectedVariables),
	}

	layout := "2006-01-02T15:04:05.000Z"
	str := "2014-11-12T11:45:26.371Z"
	timeNow, _ := time.Parse(layout, str)
	mock_time.On("Now").Return(timeNow)

	mock_reporter.On("ReportStepStart", executionId, step, cacao.NewVariables(expectedVariables), timeNow).Return()

	mock_reporter.On("ReportStepEnd", executionId, step, cacao.NewVariables(expectedVariables), nil, timeNow).Return()
	mock_ssh.On("Execute",
		metadata,
		context1).
		Return(cacao.NewVariables(expectedVariables),
			nil)

	_, err := executorObject.Execute(metadata,
		actionMetadata)

	assert.Equal(t, err, nil)
	mock_reporter.AssertExpectations(t)
	mock_ssh.AssertExpectations(t)
	mock_time.AssertExpectations(t)
}

func TestExecuteActionStep(t *testing.T) {
	mock_ssh := new(mock_capability.Mock_Capability)
	mock_http := new(mock_capability.Mock_Capability)
	mock_reporter := new(mock_reporter.Mock_Reporter)
	mock_time := new(mock_time.MockTime)

	capabilities := map[string]capability.ICapability{"ssh": mock_ssh, "http-api": mock_http}

	executorObject := New(capabilities, mock_reporter, mock_time)
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

	context1 := capability.Context{
		Command:        expectedCommand,
		Authentication: expectedAuth,
		Target:         expectedTarget,
		Variables:      cacao.NewVariables(expectedVariables),
	}

	mock_ssh.On("Execute",
		metadata,
		context1).
		Return(cacao.NewVariables(expectedVariables),
			nil)

	data := data{command: expectedCommand,
		authentication: expectedAuth,
		target:         expectedTarget,
		variables:      cacao.NewVariables(expectedVariables),
		agent:          agent}

	_, err := executorObject.executeCommands(metadata,
		data)

	assert.Equal(t, err, nil)
	mock_reporter.AssertExpectations(t)
	mock_ssh.AssertExpectations(t)
	mock_time.AssertExpectations(t)
}

func TestNonExistingCapabilityStep(t *testing.T) {
	mock_ssh := new(mock_capability.Mock_Capability)
	mock_http := new(mock_capability.Mock_Capability)
	mock_time := new(mock_time.MockTime)

	capabilities := map[string]capability.ICapability{"ssh": mock_ssh, "http-api": mock_http}

	executorObject := New(capabilities, new(mock_reporter.Mock_Reporter), mock_time)
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

	data := data{command: expectedCommand,
		authentication: expectedAuth,
		target:         expectedTarget,
		variables:      cacao.NewVariables(expectedVariables),
		agent:          agent}
	_, err := executorObject.executeCommands(metadata,
		data)

	assert.Equal(t, err, errors.New("capability: non-existing is not available in soarca"))
	mock_ssh.AssertExpectations(t)
	mock_time.AssertExpectations(t)
}

func TestVariableInterpolation(t *testing.T) {
	mock_capability1 := new(mock_capability.Mock_Capability)
	mock_time := new(mock_time.MockTime)

	capabilities := map[string]capability.ICapability{"cap1": mock_capability1}

	executorObject := New(capabilities, new(mock_reporter.Mock_Reporter), mock_time)
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

	context1 := capability.Context{Command: expectedCommand,
		Authentication: expectedAuth,
		Target:         expectedTarget,
		Variables:      cacao.NewVariables(var1, var2, var3, varUser, varPassword, varOauth, varPrivateKey, varToken, varUserId, varheader1, varheader2)}

	mock_capability1.On("Execute",
		metadata,
		context1).
		Return(cacao.NewVariables(var1),
			nil)

	data1 := data{command: inputCommand,
		authentication: inputAuth,
		target:         inputTarget,
		variables:      cacao.NewVariables(var1, var2, var3, varUser, varPassword, varOauth, varPrivateKey, varToken, varUserId, varheader1, varheader2),
		agent:          agent}

	_, err := executorObject.executeCommands(metadata,
		data1)

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
	contextHttp := capability.Context{Command: expectedHttpCommand,
		Authentication: expectedAuth,
		Target:         expectedTarget,
		Variables:      cacao.NewVariables(varHttpContent, varheader1, varheader2)}

	mock_capability1.On("Execute",
		metadataHttp,
		contextHttp).
		Return(cacao.NewVariables(var1),
			nil)

	data2 := data{command: httpCommand,
		authentication: expectedAuth,
		target:         expectedTarget,
		variables:      cacao.NewVariables(varHttpContent, varheader1, varheader2),
		agent:          agent}

	_, err = executorObject.executeCommands(metadata,
		data2)

	assert.Equal(t, err, nil)
	mock_capability1.AssertExpectations(t)
	mock_time.AssertExpectations(t)

}
