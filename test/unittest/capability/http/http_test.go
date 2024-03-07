package ssh_test

import (
	"errors"
	"soarca/internal/capability/http"
	"soarca/models/cacao"
	"testing"

	"github.com/go-playground/assert/v2"
)

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
	test_b64_content := "R08gU09BUkNBIQ=="
	expectedCommand := cacao.Command{
		Type:       "http-api",
		Command:    "GET 0.0.0.0:80/",
		ContentB64: test_b64_content,
	}

	ret_content, err := http.ObtainHttpRequestContentDataFromCommand(expectedCommand)

	assert.Equal(t, ret_content, []byte("GO SOARCA!"))
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
