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

func TestCorsHeaderFromNonAllowedOrigin(t *testing.T) {

	// Set example.com as allowed origin
	t.Setenv("SOARCA_ALLOWED_ORIGINS", "http://example.com")
	t.Setenv("PORT", "8081")

	// Start SOARCA in separate threat
	go initializeSoarca(t)

	// Wait for the server to be online
	time.Sleep(400 * time.Millisecond)

	client := http.Client{}
	buffer := bytes.NewBufferString("")
	request, err := http.NewRequest("POST", "http://localhost:8081", buffer)
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
	t.Log(response)
	assert.Equal(t, http.StatusNotFound, response.StatusCode) // We expect 404 do to the empty request body
	assert.Equal(t, "http://example.com", origins)

	client2 := http.Client{}
	buffer2 := bytes.NewBufferString("")
	request2, err := http.NewRequest("POST", "http://localhost:8081", buffer2)
	if err != nil {
		t.Fail()
	}

	request2.Header.Add("Origin", "http://example2.com")
	response2, err := client2.Do(request2)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log(response2)
	assert.Equal(t, http.StatusForbidden, response2.StatusCode)

}
