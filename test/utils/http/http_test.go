package ssh_test

import (
	"encoding/json"
	"fmt"
	"testing"

	http "soarca/utils/http"

	"github.com/go-playground/assert/v2"
)

type bearerResponse struct {
	Authenticated bool   `json:"authenticated"`
	Token         string `json:"token"`
}
type basicAuthReponse struct {
	Authenticated bool   `json:"authenticated"`
	User          string `json:"user"`
}

type testJson struct {
	Id          string `json:"id"`
	User        string `json:"user"`
	Description string `json:"description"`
}

type httpBinResponseBody struct {
	Data string `json:"data"`
}

// Test general http options, we do not check responses body, as these are variable for the general connection tests
func TestHttpGetConnection(t *testing.T) {
	httpRequest := http.HttpRequest{}
	httpOptions := http.HttpOptions{
		Method:  "GET",
		Url:     "https://httpbin.org/get",
		Headers: map[string]string{"accept": "application/json"},
	}
	response, err := httpRequest.Request(httpOptions)
	t.Log(string(response))
	if err != nil {
		t.Fatalf(fmt.Sprint("http get request test has failed: ", err))
	}
	if len(response) == 0 {
		t.Fatalf("empty response")
	}
	t.Log(string(response))
}

func TestHttpPostConnection(t *testing.T) {
	httpRequest := http.HttpRequest{}
	httpOptions := http.HttpOptions{
		Method:  "POST",
		Url:     "https://httpbin.org/post",
		Headers: map[string]string{"accept": "application/json"},
	}
	response, err := httpRequest.Request(httpOptions)
	t.Log(string(response))
	if err != nil {
		t.Fatalf(fmt.Sprint("http post request test has failed: ", err))
	}
	if len(response) == 0 {
		t.Fatalf("empty response")
	}
	t.Log(string(response))
}

func TestHttpPutConnection(t *testing.T) {
	httpRequest := http.HttpRequest{}
	httpOptions := http.HttpOptions{
		Method:  "PUT",
		Url:     "https://httpbin.org/put",
		Headers: map[string]string{"accept": "application/json"},
	}
	response, err := httpRequest.Request(httpOptions)
	if err != nil {
		t.Fatalf(fmt.Sprint("http put request test has failed: ", err))
	}
	if len(response) == 0 {
		t.Fatalf("empty response")
	}
	t.Log(string(response))
}

func TestHttpDeleteConnection(t *testing.T) {
	httpRequest := http.HttpRequest{}
	httpOptions := http.HttpOptions{
		Method:  "DELETE",
		Url:     "https://httpbin.org/delete",
		Headers: map[string]string{"accept": "application/json"},
	}
	response, err := httpRequest.Request(httpOptions)
	if err != nil {
		t.Fatalf(fmt.Sprint("http delete request test has failed: ", err))
	}
	if len(response) == 0 {
		t.Fatalf("empty response")
	}
	t.Log(string(response))
}

// test status codes handling

func TestHttpStatus200(t *testing.T) {
	httpRequest := http.HttpRequest{}
	httpOptions := http.HttpOptions{
		Method:  "GET",
		Url:     "https://httpbin.org/status/200",
		Headers: map[string]string{"accept": "application/json"},
	}
	// error codes are handled internally by request, with throw error != in 200 range
	_, err := httpRequest.Request(httpOptions)
	t.Log(err)
	if err != nil {
		t.Fatalf(fmt.Sprint("http get request test has failed: ", err))
		return
	}
}

func TestHttpBearerToken(t *testing.T) {
	bearerToken := "test"
	httpRequest := http.HttpRequest{}
	httpOptions := http.HttpOptions{
		Method:  "GET",
		Url:     "https://httpbin.org/bearer",
		Headers: map[string]string{"accept": "application/json"},
		Token:   bearerToken,
	}
	response, err := httpRequest.Request(httpOptions)
	if err != nil {
		t.Fatalf(fmt.Sprint("http request has failed: ", err))
	}
	if len(response) == 0 {
		t.Fatalf("empty response")
	}
	t.Log(string(response))

	var responseToken bearerResponse
	err = json.Unmarshal(response, &responseToken)
	if err != nil {
		fmt.Println("error decoding JSON: ", err)
		t.Fatalf(fmt.Sprint("could not unmashall reponsetoken for bearer test: ", err))
	}
	assert.Equal(t, responseToken.Authenticated, true)
	assert.Equal(t, responseToken.Token, bearerToken)
}

func TestHttpBasicAuth(t *testing.T) {
	username := "test"
	password := "password"
	url := fmt.Sprintf("https://httpbin.org/basic-auth/%s/%s", username, password)
	httpRequest := http.HttpRequest{}
	httpOptions := http.HttpOptions{
		Method:   "GET",
		Url:      url,
		Headers:  map[string]string{"accept": "application/json"},
		Username: username,
		Password: password,
	}
	response, err := httpRequest.Request(httpOptions)
	if err != nil {
		t.Fatalf(fmt.Sprint("http auth Request has failed: ", err))
	}
	if len(response) == 0 {
		t.Fatalf("empty response")
	}
	t.Log(string(response))
	var authResponse basicAuthReponse
	err = json.Unmarshal(response, &authResponse)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		t.Fatalf(fmt.Sprint("could not unmashall reponse token for bearer test,", err))
	}

	assert.Equal(t, authResponse.Authenticated, true)
	assert.Equal(t, authResponse.User, username)
}

func TestHttpPostWithContentConnection(t *testing.T) {
	httpRequest := http.HttpRequest{}

	testJsonObj := testJson{Id: "28818819", User: "test", Description: "very interesting description"}
	requestBody, err := json.Marshal(testJsonObj)
	if err != nil {
		fmt.Println("error Marshall JSON:", err)
	}
	if len(requestBody) == 0 {
		t.Fatalf("empty response")
	}
	httpOptions := http.HttpOptions{
		Method:  "POST",
		Url:     "https://httpbin.org/anything",
		Headers: map[string]string{"accept": "application/json"},
		Body:    requestBody,
	}
	response, err := httpRequest.Request(httpOptions)

	t.Log(string(response))
	if err != nil {
		t.Fatalf(fmt.Sprint("http post request with body content has failed: ", err))
	}

	var httpBinReponse httpBinResponseBody
	err = json.Unmarshal(response, &httpBinReponse)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		t.Fatalf(fmt.Sprint("Could not unmashall reponse token for bearer test,", err))
	}

	var responseToken bearerResponse
	err = json.Unmarshal([]byte(httpBinReponse.Data), &responseToken)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		t.Fatalf(fmt.Sprint("Could not unmashall reponse token for bearer test,", err))
	}
	assert.Equal(t, testJsonObj.Id, "28818819")
	assert.Equal(t, testJsonObj.User, "test")
	assert.Equal(t, testJsonObj.Description, "very interesting description")
}
