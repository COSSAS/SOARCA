package http_test

// Build http capability with New() using mock http Request
// test correct parsing of HttpOptions fields and errors handling

import (
	"errors"
	"soarca/internal/capability/http"
	"soarca/models/cacao"
	"soarca/models/execution"
	mock_request "soarca/test/unittest/mocks/mock_utils/http"
	http_request "soarca/utils/http"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestHTTPOptionsCorrectlyGenerated(t *testing.T) {
	mock_http_request := new(mock_request.MockHttpRequest)
	httpCapability := http.New(mock_http_request)

	var oauth2_info = cacao.AuthenticationInformation{
		ID:    "6ba7b810-9dad-11d1-80b4-00c04fd430c9",
		Type:  "oauth2",
		Token: "this-is-a-test",
	}

	target := cacao.AgentTarget{Address: map[cacao.NetAddressType][]string{
		"url": {"https://httpbin.org/post"},
	}}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}

	var variable1 = cacao.Variable{
		Type:  "string",
		Name:  "test request building",
		Value: "",
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}

	httpOptions := http_request.HttpOptions{
		Command: &command,
		Target:  &target,
		Auth:    &oauth2_info,
	}

	payload := "payload test"
	payload_byte := []byte(payload)
	mock_http_request.On("Request", httpOptions).Return(payload_byte, nil)

	outArg := cacao.Variable{
		Type: cacao.VariableTypeString,
		Name: "__soarca_http_api_result__",
	}
	inArgs := []string{variable1.Name}
	outArgs := []string{outArg.Name}

	results, err := httpCapability.Execute(
		metadata,
		command,
		oauth2_info,
		target,
		cacao.NewVariables(variable1, outArg),
		inArgs,
		outArgs)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, results["__soarca_http_api_result__"].Value, payload)
	t.Log(results)

	mock_http_request.AssertExpectations(t)
}

func TestHTTPOptionsEmptyAuth(t *testing.T) {
	mock_http_request := &mock_request.MockHttpRequest{}
	httpCapability := http.New(mock_http_request)

	target := cacao.AgentTarget{Address: map[cacao.NetAddressType][]string{
		"url": {"https://httpbin.org/post"},
	}}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}

	var variable1 = cacao.Variable{
		Type:  "string",
		Name:  "test request building",
		Value: "",
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}
	empty_auth := new(cacao.AuthenticationInformation)

	httpOptions := http_request.HttpOptions{
		Command: &command,
		Target:  &target,
		Auth:    empty_auth,
	}

	payload := "payload test"
	payload_byte := []byte(payload)
	mock_http_request.On("Request", httpOptions).Return(payload_byte, nil)

	outArg := cacao.Variable{
		Type: cacao.VariableTypeString,
		Name: "__soarca_http_api_result__",
	}
	inArgs := []string{variable1.Name}
	outArgs := []string{outArg.Name}

	results, err := httpCapability.Execute(
		metadata,
		command,
		*empty_auth,
		target,
		cacao.NewVariables(variable1, outArg),
		inArgs,
		outArgs)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, results["__soarca_http_api_result__"].Value, payload)
	t.Log(results)

	mock_http_request.AssertExpectations(t)
}

func TestHTTPOptionsEmptyCommand(t *testing.T) {
	mock_http_request := &mock_request.MockHttpRequest{}
	httpCapability := http.New(mock_http_request)

	target := cacao.AgentTarget{Address: map[cacao.NetAddressType][]string{
		"url": {"https://httpbin.org/post"},
	}}
	empty_command := new(cacao.Command)

	var oauth2_info = cacao.AuthenticationInformation{
		ID:    "6ba7b810-9dad-11d1-80b4-00c04fd430c9",
		Type:  "oauth2",
		Token: "this-is-a-test",
	}

	var variable1 = cacao.Variable{
		Type:  "string",
		Name:  "test request building",
		Value: "",
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}

	httpOptions := http_request.HttpOptions{
		Command: empty_command,
		Target:  &target,
		Auth:    &oauth2_info,
	}

	expected_error := errors.New("command pointer is empty")
	mock_http_request.On("Request", httpOptions).Return([]byte{}, expected_error)

	outArg := cacao.Variable{
		Type: cacao.VariableTypeString,
		Name: "__soarca_http_api_result__",
	}
	inArgs := []string{variable1.Name}
	outArgs := []string{outArg.Name}

	results, err := httpCapability.Execute(
		metadata,
		*empty_command,
		oauth2_info,
		target,
		cacao.NewVariables(variable1, outArg),
		inArgs,
		outArgs)
	if err == nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, err, expected_error)
	t.Log(results)

	mock_http_request.AssertExpectations(t)
}

func TestNoOutArgs(t *testing.T) {
	mock_http_request := new(mock_request.MockHttpRequest)
	httpCapability := http.New(mock_http_request)

	var oauth2_info = cacao.AuthenticationInformation{
		ID:    "6ba7b810-9dad-11d1-80b4-00c04fd430c9",
		Type:  "oauth2",
		Token: "this-is-a-test",
	}

	target := cacao.AgentTarget{Address: map[cacao.NetAddressType][]string{
		"url": {"https://httpbin.org/post"},
	}}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}

	var variable1 = cacao.Variable{
		Type:  "string",
		Name:  "test request building",
		Value: "",
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}

	httpOptions := http_request.HttpOptions{
		Command: &command,
		Target:  &target,
		Auth:    &oauth2_info,
	}

	payload := "payload test"
	payload_byte := []byte(payload)
	mock_http_request.On("Request", httpOptions).Return(payload_byte, nil)

	inArgs := []string{variable1.Name}
	outArgs := []string{}

	results, err := httpCapability.Execute(
		metadata,
		command,
		oauth2_info,
		target,
		cacao.NewVariables(variable1),
		inArgs,
		outArgs)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(results), 0)
	t.Log(results)

	mock_http_request.AssertExpectations(t)
}

func TestToManyOutArgs(t *testing.T) {
	mock_http_request := new(mock_request.MockHttpRequest)
	httpCapability := http.New(mock_http_request)

	var oauth2_info = cacao.AuthenticationInformation{
		ID:    "6ba7b810-9dad-11d1-80b4-00c04fd430c9",
		Type:  "oauth2",
		Token: "this-is-a-test",
	}

	target := cacao.AgentTarget{Address: map[cacao.NetAddressType][]string{
		"url": {"https://httpbin.org/post"},
	}}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}

	var variable1 = cacao.Variable{
		Type:  "string",
		Name:  "test request building",
		Value: "",
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}

	httpOptions := http_request.HttpOptions{
		Command: &command,
		Target:  &target,
		Auth:    &oauth2_info,
	}

	payload := "payload test"
	payload_byte := []byte(payload)
	mock_http_request.On("Request", httpOptions).Return(payload_byte, nil)

	outArg := cacao.Variable{
		Type: cacao.VariableTypeString,
		Name: "__soarca_http_api_result__",
	}
	outArg2 := cacao.Variable{
		Type: cacao.VariableTypeString,
		Name: "__soarca_http_api_result2__",
	}
	inArgs := []string{variable1.Name}
	outArgs := []string{outArg.Name, outArg2.Name}

	results, err := httpCapability.Execute(
		metadata,
		command,
		oauth2_info,
		target,
		cacao.NewVariables(variable1, outArg, outArg2),
		inArgs,
		outArgs)
	assert.Equal(t, err, errors.New("number of outargs does not match data array length"))
	assert.Equal(t, len(results), 0)
	t.Log(results)

	mock_http_request.AssertExpectations(t)
}

func TestOutArgNotFound(t *testing.T) {
	mock_http_request := new(mock_request.MockHttpRequest)
	httpCapability := http.New(mock_http_request)

	var oauth2_info = cacao.AuthenticationInformation{
		ID:    "6ba7b810-9dad-11d1-80b4-00c04fd430c9",
		Type:  "oauth2",
		Token: "this-is-a-test",
	}

	target := cacao.AgentTarget{Address: map[cacao.NetAddressType][]string{
		"url": {"https://httpbin.org/post"},
	}}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}

	var variable1 = cacao.Variable{
		Type:  "string",
		Name:  "test request building",
		Value: "",
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}

	httpOptions := http_request.HttpOptions{
		Command: &command,
		Target:  &target,
		Auth:    &oauth2_info,
	}

	payload := "payload test"
	payload_byte := []byte(payload)
	mock_http_request.On("Request", httpOptions).Return(payload_byte, nil)

	outArg := cacao.Variable{
		Type: cacao.VariableTypeString,
		Name: "__soarca_http_api_result__",
	}
	outArg2 := cacao.Variable{
		Type: cacao.VariableTypeString,
		Name: "__soarca_http_api_result2__",
	}
	inArgs := []string{variable1.Name}
	outArgs := []string{"__1__"}

	results, err := httpCapability.Execute(
		metadata,
		command,
		oauth2_info,
		target,
		cacao.NewVariables(variable1, outArg, outArg2),
		inArgs,
		outArgs)
	assert.Equal(t, err, errors.New("key is not found in variables"))
	assert.Equal(t, len(results), 0)
	t.Log(results)

	mock_http_request.AssertExpectations(t)
}
