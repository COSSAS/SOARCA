package api_test

import (
	"bytes"
	"net/http"
	"soarca/internal/controller"
	"testing"

	"github.com/stretchr/testify/assert"
)

func initializeSoarca(t *testing.T) {
	err := controller.Initialize()
	if err != nil {
		t.Log(err)
	}
}

func TestCorsHeader(t *testing.T) {
	go initializeSoarca(t)

	client := http.Client{}
	buffer := bytes.NewBufferString("")
	request, err := http.NewRequest("POST", "http://localhost:8080", buffer)
	if err != nil {
		t.Fail()
	}

	request.Header.Add("Origin", "http://example.com")
	response, _ := client.Do(request)
	origins := response.Header.Get("Access-Control-Allow-Origin")
	assert.Equal(t, "*", origins)

}
