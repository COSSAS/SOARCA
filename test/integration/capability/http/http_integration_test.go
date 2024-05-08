package http_integrations_test

import (
	"fmt"
	"testing"

	"soarca/internal/capability/http"
	"soarca/models/cacao"
	"soarca/models/execution"
	httpUtil "soarca/utils/http"

	http "soarca/utils/http"

	"github.com/go-playground/assert"
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
	expectedCommand := cacao.Command{
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
	// But what to do if there is no target and no AuthInfo?
	results, err := httpCapability.Execute(
		metadata, expectedCommand,
		cacao.AuthenticationInformation{},
		target,
		cacao.NewVariables(variable1))
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
	command := cacao.Command{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}

	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId, _ := uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	stepId, _ := uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}
	results, err := httpCapability.Execute(
		metadata,
		command,
		auth,
		target,
		cacao.NewVariables())
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(results)
}

func TestHttpBasicAuth(t *testing.T) {
	request := httpUtil.HttpRequest{}
	httpCapability := http.New(&request)
	username := "test"
	password := "password"
	url := fmt.Sprintf("https://httpbin.org/basic-auth/%s/%s", username, password)

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {url},
		},
		AuthInfoIdentifier: "d0c7e6a0-f7fe-464e-9935-e6b3443f5b91",
	}

	auth := cacao.AuthenticationInformation{
		Type:     cacao.AuthInfoHTTPBasicType,
		Username: username,
		Password: password,
		ID:       "d0c7e6a0-f7fe-464e-9935-e6b3443f5b91",
	}

	command := cacao.Command{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId, _ := uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	stepId, _ := uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}
	results, err := httpCapability.Execute(
		metadata,
		command,
		auth,
		target,
		cacao.NewVariables())
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(results)
}

func TestInsecureHTTPConnection(t *testing.T) {
	httpRequest := http.HttpRequest{}

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://localhost/get"},
		},
	}
	command := cacao.Command{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := http.HttpOptions{
		Command: &command,
		Target:  &target,
	}
	httpRequest.SkipCertificateValidation(true)
	response, err := httpRequest.Request(httpOptions)
	t.Log(string(response))
	if err != nil {
		t.Error("http get request test has failed: ", err)
	}
	if len(response) == 0 {
		t.Error("empty response")
	}
	t.Log(string(response))
	assert.Equal(t, err, nil)
}

func TestInsecureHTTPConnectionWithFailure(t *testing.T) {
	httpRequest := http.HttpRequest{}

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://localhost/get"},
		},
	}
	command := cacao.Command{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := http.HttpOptions{
		Command: &command,
		Target:  &target,
	}

	response, err := httpRequest.Request(httpOptions)
	t.Log(string(response))
	if err == nil {
		t.Error("test should have failed as insecure is not allowed")
	}
	assert.NotEqual(t, err, nil)
}
