package api_test

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTest(t *testing.T) {
	// Start SOARCA in separate threat
	t.Setenv("PORT", "8084")
	go initializeSoarca(t)

	// Wait for the server to be online
	time.Sleep(400 * time.Millisecond)

	client := http.Client{}
	buffer := bytes.NewBufferString("")
	request, err := http.NewRequest("POST", "http://localhost:8084/trigger/", buffer)
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
