package http

// Build http capability with New() using mock http Request
// test correct parsing of HttpOptions fields and errors handling

import (
	"errors"

	"soarca/internal/workflow/capability"
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
	http_request "soarca/pkg/utils/http"
	mock_request "soarca/test/unittest/mocks/mock_utils/http"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestHTTPOptionsCorrectlyGenerated(t *testing.T) {
	mock_http_request := new(mock_request.MockHttpRequest)
	httpCapability := New(mock_http_request)

	var oauth2_info = cacao.AuthenticationInformation{
		ID:    "6ba7b810-9dad-11d1-80b4-00c04fd430c9",
		Type:  "oauth2",
		Token: "this-is-a-test",
	}

	target := cacao.AgentTarget{Address: map[cacao.NetAddressType][]string{
		"url": {"https://localhost/post"},
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

	var runId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := run.Metadata{RunId: runId, PlaybookId: playbookId.String(), StepId: stepId.String()}

	httpOptions := http_request.HttpOptions{
		Command: &command,
		Target:  &target,
		Auth:    &oauth2_info,
	}

	payload := "payload test"
	payload_byte := []byte(payload)
	mock_http_request.On("Request", httpOptions).Return(payload_byte, nil)

	data := capability.Context{
		Commands:  []cacao.Command{command},
		Targets:   []capability.ResolvedTarget{{Target: target, Authentication: oauth2_info}},
		Variables: cacao.NewVariables(variable1),
	}

	results, err := httpCapability.Execute(
		metadata,
		data)
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
	httpCapability := New(mock_http_request)

	target := cacao.AgentTarget{Address: map[cacao.NetAddressType][]string{
		"url": {"https://localhost/post"},
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

	var runId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := run.Metadata{RunId: runId, PlaybookId: playbookId.String(), StepId: stepId.String()}
	empty_auth := new(cacao.AuthenticationInformation)

	httpOptions := http_request.HttpOptions{
		Command: &command,
		Target:  &target,
		Auth:    empty_auth,
	}

	payload := "payload test"
	payload_byte := []byte(payload)
	mock_http_request.On("Request", httpOptions).Return(payload_byte, nil)

	data := capability.Context{
		Commands:  []cacao.Command{command},
		Targets:   []capability.ResolvedTarget{{Target: target, Authentication: *empty_auth}},
		Variables: cacao.NewVariables(variable1),
	}

	results, err := httpCapability.Execute(
		metadata,
		data)
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
	httpCapability := New(mock_http_request)

	target := cacao.AgentTarget{Address: map[cacao.NetAddressType][]string{
		"url": {"https://localhost/post"},
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

	var runId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := run.Metadata{RunId: runId, PlaybookId: playbookId.String(), StepId: stepId.String()}

	httpOptions := http_request.HttpOptions{
		Command: empty_command,
		Target:  &target,
		Auth:    &oauth2_info,
	}

	expected_error := errors.New("command pointer is empty")
	mock_http_request.On("Request", httpOptions).Return([]byte{}, expected_error)

	data := capability.Context{
		Commands:  []cacao.Command{*empty_command},
		Targets:   []capability.ResolvedTarget{{Target: target, Authentication: oauth2_info}},
		Variables: cacao.NewVariables(variable1),
	}

	results, err := httpCapability.Execute(
		metadata,
		data)
	if err == nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, err, expected_error)
	t.Log(results)

	mock_http_request.AssertExpectations(t)
}

func TestHTTPExecuteNoTargetsSkipsWithoutError(t *testing.T) {
	mock_http_request := new(mock_request.MockHttpRequest)
	httpCapability := New(mock_http_request)

	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
	}

	var runId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := run.Metadata{RunId: runId, PlaybookId: playbookId.String(), StepId: stepId.String()}

	data := capability.Context{
		Commands: []cacao.Command{command},
		Targets:  []capability.ResolvedTarget{},
	}

	results, err := httpCapability.Execute(metadata, data)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, len(results), 0)

	mock_http_request.AssertNotCalled(t, "Request")
}
