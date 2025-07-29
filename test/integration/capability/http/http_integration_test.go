package http_integrations_test

import (
	"fmt"
	"testing"

	"soarca/pkg/core/capability"
	"soarca/pkg/core/capability/http"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	httpUtil "soarca/pkg/utils/http"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestHttpConnection(t *testing.T) {
	request := httpUtil.HttpRequest{}
	httpCapability := http.New(&request)

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://httpbin.org/get"},
		},
	}
	expectedCommand := cacao.CommandData{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}

	variable1 := cacao.Variable{
		Type:  "string",
		Name:  "test_auth",
		Value: "",
	}

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId, _ := uuid.Parse("playbook--d09351a2-a075-40c8-8054-0b7c423db83f")
	stepId, _ := uuid.Parse("action--81eff59f-d084-4324-9e0a-59e353dbd28f")

	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}
	data := capability.Context{
		CommandData: expectedCommand,
		Target:      target,
		Variables:   cacao.NewVariables(variable1),
	}
	// But what to do if there is no target and no AuthInfo?
	results, err := httpCapability.Execute(
		metadata, data)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	fmt.Println(results)
}

func TestHttpOAuth2(t *testing.T) {
	request := httpUtil.HttpRequest{}
	httpCapability := http.New(&request)

	bearerToken := "test_token"

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://httpbin.org/bearer"},
		},
		AuthInfoIdentifier: "d0c7e6a0-f7fe-464e-9935-e6b3443f5b91",
	}
	auth := cacao.AuthenticationInformation{
		Type:  cacao.AuthInfoOAuth2Type,
		Token: bearerToken,
		ID:    "d0c7e6a0-f7fe-464e-9935-e6b3443f5b91",
	}
	command := cacao.CommandData{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId, _ := uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	stepId, _ := uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}
	data := capability.Context{
		CommandData:    command,
		Target:         target,
		Authentication: auth,
		Variables:      cacao.NewVariables(),
	}
	results, err := httpCapability.Execute(
		metadata,
		data)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(results)
}

func TestHttpBasicAuth(t *testing.T) {
	request := httpUtil.HttpRequest{}
	httpCapability := http.New(&request)
	user_id := "test"
	password := "password"
	url := fmt.Sprintf("https://httpbin.org/basic-auth/%s/%s", user_id, password)

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {url},
		},
		AuthInfoIdentifier: "d0c7e6a0-f7fe-464e-9935-e6b3443f5b91",
	}

	auth := cacao.AuthenticationInformation{
		Type:     cacao.AuthInfoHTTPBasicType,
		UserId:   user_id,
		Password: password,
		ID:       "d0c7e6a0-f7fe-464e-9935-e6b3443f5b91",
	}

	command := cacao.CommandData{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId, _ := uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	stepId, _ := uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}
	data := capability.Context{
		CommandData:    command,
		Target:         target,
		Authentication: auth,
		Variables:      cacao.NewVariables(),
	}
	results, err := httpCapability.Execute(
		metadata,
		data)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(results)
}

func TestInsecureHTTPConnection(t *testing.T) {
	httpRequest := httpUtil.HttpRequest{}

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://localhost/get"},
		},
	}
	command := cacao.CommandData{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := httpUtil.HttpOptions{
		Command: &command,
		Target:  &target,
	}
	httpRequest.SkipCertificateValidation(true)
	response, err := httpRequest.Request(httpOptions)
	assert.Equal(t, err, nil)
	t.Log(string(response))
	if len(response) == 0 {
		t.Error("empty response")
	}
	t.Log(string(response))
}

func TestInsecureHTTPConnectionWithFailure(t *testing.T) {
	httpRequest := httpUtil.HttpRequest{}

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://localhost/get"},
		},
	}
	command := cacao.CommandData{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := httpUtil.HttpOptions{
		Command: &command,
		Target:  &target,
	}

	response, err := httpRequest.Request(httpOptions)
	assert.NotEqual(t, err, nil)
	t.Log(string(response))
}
