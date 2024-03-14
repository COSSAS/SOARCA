package openc2_test

import (
	"testing"

	openc2 "soarca/internal/capability/openc2"
	"soarca/models/cacao"
	"soarca/models/execution"
	mockRequest "soarca/test/unittest/mocks/mock_utils/http"
	"soarca/utils/http"

	assert "github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestOpenC2Request(t *testing.T) {
	mockHttp := &mockRequest.MockHttpRequest{}
	openc2 := openc2.New(mockHttp)

	authId, _ := uuid.Parse("6aa7b810-9dad-11d1-81b4-00c04fd430c8")
	executionId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId, _ := uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	stepId, _ := uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://soarca.tno.nl"},
		},
		AuthInfoIdentifier: authId.String(),
	}

	auth := cacao.AuthenticationInformation{
		ID:    authId.String(),
		Type:  "oauth2",
		Token: "this-is-a-test",
	}

	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}

	cacaoVariable := cacao.Variable{
		Type:  "string",
		Name:  "test request building",
		Value: "",
	}

	metadata := execution.Metadata{
		ExecutionId: executionId,
		PlaybookId:  playbookId.String(),
		StepId:      stepId.String(),
	}

	httpOptions := http.HttpOptions{
		Command: &command,
		Target:  &target,
		Auth:    &auth,
	}

	payload := "test payload"

	payloadBytes := []byte(payload)

	mockHttp.On("Request", httpOptions).Return(payloadBytes, nil)

	results, err := openc2.Execute(
		metadata,
		command,
		auth,
		target,
		cacao.NewVariables(cacaoVariable))
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(results)
	assert.Equal(t, results["__soarca_openc2_http_result__"].Value, payload)
}
