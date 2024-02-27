package ssh_test

import (
	"encoding/hex"
	"errors"
	"fmt"
	"soarca/internal/capability/http"
	"soarca/models/cacao"
	"soarca/models/execution"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestHttpConnection(t *testing.T) {
	httpCapability := new(http.HttpCapability)

	expectedCommand := cacao.Command{
		Type:    "http-api",
		Command: "GET https://httpbin.org/",
		Headers: map[string]string{"accept": "application/json"},
	}

	var variable1 = cacao.Variable{
		Type:  "string",
		Name:  "test_auth",
		Value: "",
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("playbook--d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("action--81eff59f-d084-4324-9e0a-59e353dbd28f")

	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}
	// But what to do if there is no target and no AuthInfo?
	results, err := httpCapability.Execute(
		metadata, expectedCommand,
		cacao.AuthenticationInformation{},
		cacao.AgentTarget{},
		map[string]cacao.Variable{"test": variable1})
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	fmt.Println(results)
}

func TestHttpOAuth2(t *testing.T) {
	httpCapability := new(http.HttpCapability)

	var oauth2_info = cacao.AuthenticationInformation{
		ID:    "6ba7b810-9dad-11d1-80b4-00c04fd430c9",
		Type:  "oauth2",
		Token: "this-is-a-test",
	}

	expectedCommand := cacao.Command{
		Type:    "http-api",
		Command: "GET https://httpbin.org/bearer",
		Headers: map[string]string{"accept": "application/json"},
	}

	var variable1 = cacao.Variable{
		Type:  "string",
		Name:  "test_auth",
		Value: "",
	}

	var target = cacao.AgentTarget{
		ID:                 "6ba7b810-9dad-11d1-80b4-00c04fd430c0",
		Type:               "http-api",
		Name:               "Cybersec APIs",
		AuthInfoIdentifier: "6ba7b810-9dad-11d1-80b4-00c04fd430c9",
	}

	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}
	results, err := httpCapability.Execute(
		metadata,
		expectedCommand,
		oauth2_info,
		target,
		map[string]cacao.Variable{"test": variable1})
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(results)
}

func TestHttpBasicAuth(t *testing.T) {
	httpCapability := new(http.HttpCapability)

	var basicauth_info = cacao.AuthenticationInformation{
		ID:       "6ba7b810-9dad-11d1-80b4-00c04fd430c9",
		Type:     "http-basic",
		Username: "username_test",
		Password: "password_test",
	}

	expectedCommand := cacao.Command{
		Type:    "http-api",
		Command: "GET https://httpbin.org/hidden-basic-auth/username_test/password_test",
		Headers: map[string]string{"accept": "application/json"},
	}

	var variable1 = cacao.Variable{
		Type:  "string",
		Name:  "test_auth",
		Value: "",
	}

	var target = cacao.AgentTarget{
		ID:                 "6ba7b810-9dad-11d1-80b4-00c04fd430c0",
		Type:               "http-api",
		Name:               "Cybersec APIs",
		AuthInfoIdentifier: "6ba7b810-9dad-11d1-80b4-00c04fd430c9",
	}
	var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var playbookId, _ = uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	var stepId, _ = uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")
	metadata := execution.Metadata{ExecutionId: executionId, PlaybookId: playbookId.String(), StepId: stepId.String()}
	results, err := httpCapability.Execute(
		metadata,
		expectedCommand,
		basicauth_info,
		target, map[string]cacao.Variable{"test": variable1})
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(results)
}

// Tests for data fetching from command
func TestHttpObtainMethodFromCommandValid(t *testing.T) {

	expectedCommand := cacao.Command{
		Type:    "http-api",
		Command: "POST https://google.com/",
	}

	httpMethod, httpUrl, err := http.ObtainHttpMethodAndUrlFromCommand(expectedCommand)
	assert.Equal(t, httpMethod, "POST")
	assert.Equal(t, httpUrl, "https://google.com/")
	assert.Equal(t, err, nil)
}

func TestHttpObtainMethodAndUrlFromCommandInvalid(t *testing.T) {

	expectedCommand := cacao.Command{
		Type:    "http-api",
		Command: "https://google.com/", // No method
	}

	httpMethod, httpUrl, err := http.ObtainHttpMethodAndUrlFromCommand(expectedCommand)

	assert.Equal(t, httpMethod, "")
	assert.Equal(t, httpUrl, "")
	assert.Equal(t, err, errors.New("method or url missing from command"))

}

// Tests obtain content from command
func TestObtainHttpRequestContentDataFromCommandBothTypes(t *testing.T) {
	test_content := "414141"
	test_b64_content := "923948a09a"
	expectedCommand := cacao.Command{
		Type:       "http-api",
		Command:    "GET 0.0.0.0:80/",
		Content:    test_content,
		ContentB64: test_b64_content,
	}

	ret_content, err := http.ObtainHttpRequestContentDataFromCommand(expectedCommand)

	assert.Equal(t, ret_content, []byte(test_content))
	assert.Equal(t, err, nil)
}
func TestObtainHttpRequestContentDataFromCommandB64Only(t *testing.T) {
	test_b64_content := "923948a09a"
	expectedCommand := cacao.Command{
		Type:       "http-api",
		Command:    "GET 0.0.0.0:80/",
		ContentB64: test_b64_content,
	}

	ret_content, err := http.ObtainHttpRequestContentDataFromCommand(expectedCommand)

	expected, _ := hex.DecodeString(test_b64_content)
	assert.Equal(t, ret_content, expected)
	assert.Equal(t, err, nil)
}
func TestObtainHttpRequestContentDataFromCommandPlainTextOnly(t *testing.T) {
	test_content := "414141"
	expectedCommand := cacao.Command{
		Type:    "http-api",
		Command: "GET 0.0.0.0:80/",
		Content: test_content,
	}

	ret_content, err := http.ObtainHttpRequestContentDataFromCommand(expectedCommand)

	assert.Equal(t, ret_content, []byte(test_content))
	assert.Equal(t, err, nil)
}

func TestObtainHttpRequestContentDataFromCommandEmpty(t *testing.T) {
	expectedCommand := cacao.Command{
		Type:    "http-api",
		Command: "GET 0.0.0.0:80/",
	}

	ret_content, err := http.ObtainHttpRequestContentDataFromCommand(expectedCommand)

	assert.Equal(t, ret_content, nil)
	assert.Equal(t, err, nil)
}
