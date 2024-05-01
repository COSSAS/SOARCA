package api_test

import (
	"bytes"
	"net/http"
	"soarca/internal/controller"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func initializeSoarca(t *testing.T) {
	err := controller.Initialize()
	if err != nil {
		t.Log(err)
	}
}

func TestCorsHeader(t *testing.T) {
	// Start SOARCA in separate threat
	go initializeSoarca(t)

	// Wait for the server to be online
	time.Sleep(400 * time.Millisecond)

	client := http.Client{}
	buffer := bytes.NewBufferString("")
	request, err := http.NewRequest("POST", "http://localhost:8080", buffer)
	if err != nil {
		t.Fail()
	}

	request.Header.Add("Origin", "http://example.com")
	response, err := client.Do(request)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	origins := response.Header.Get("Access-Control-Allow-Origin")
	assert.Equal(t, "*", origins)

}
