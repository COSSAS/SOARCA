package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"soarca/internal/controller"
	"soarca/models/api"
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

func TestPingPong(t *testing.T) {
	// Start SOARCA in separate threat
	t.Setenv("PORT", "8082")
	go initializeSoarca(t)

	// Wait for the server to be online
	time.Sleep(400 * time.Millisecond)

	client := http.Client{}
	buffer := bytes.NewBufferString("")
	request, err := http.NewRequest("GET", "http://localhost:8082/status/ping", buffer)
	if err != nil {
		t.Fail()
	}

	// request.Header.Add("Origin", "http://example.com")
	response, err := client.Do(request)
	t.Log(response)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, response.StatusCode)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	assert.Equal(t, nil, err)

	assert.Equal(t, "pong", string(body))

}

func TestStatus(t *testing.T) {
	// Start SOARCA in separate threat
	t.Setenv("PORT", "8083")
	go initializeSoarca(t)

	// Wait for the server to be online
	time.Sleep(400 * time.Millisecond)

	client := http.Client{}
	buffer := bytes.NewBufferString("")
	request, err := http.NewRequest("GET", "http://localhost:8083/status", buffer)
	if err != nil {
		t.Fail()
	}

	response, err := client.Do(request)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	assert.Equal(t, http.StatusOK, response.StatusCode)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	assert.Equal(t, nil, err)

	status := api.Status{}
	err = json.Unmarshal(body, &status)
	assert.Equal(t, nil, err)

	assert.Equal(t, "production", status.Mode)
	assert.NotEmpty(t, status.Mode)
	assert.NotEmpty(t, status.Runtime)
	assert.NotEmpty(t, status.Time)
	t.Log(status)

}
