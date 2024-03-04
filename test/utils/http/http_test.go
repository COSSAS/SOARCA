package ssh_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"soarca/models/cacao"
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

	target := cacao.AgentTarget{HttpUrl: "https://httpbin.org/get"}
	httpOptions := http.HttpOptions{
		Method:  "GET",
		Target:  target,
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

	target := cacao.AgentTarget{HttpUrl: "https://httpbin.org/post"}
	httpOptions := http.HttpOptions{
		Method:  "POST",
		Target:  target,
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
	target := cacao.AgentTarget{HttpUrl: "https://httpbin.org/put"}
	httpOptions := http.HttpOptions{
		Method:  "PUT",
		Target:  target,
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
	target := cacao.AgentTarget{HttpUrl: "https://httpbin.org/delete"}
	httpOptions := http.HttpOptions{
		Method:  "DELETE",
		Target:  target,
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
	target := cacao.AgentTarget{HttpUrl: "https://httpbin.org/status/200"}
	httpOptions := http.HttpOptions{
		Method:  "GET",
		Target:  target,
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
	target := cacao.AgentTarget{HttpUrl: "https://httpbin.org/bearer"}
	httpOptions := http.HttpOptions{
		Method:  "GET",
		Target:  target,
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

	target := cacao.AgentTarget{HttpUrl: url}

	httpOptions := http.HttpOptions{
		Method:   "GET",
		Target:   target,
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
	target := cacao.AgentTarget{HttpUrl: "https://httpbin.org/anything"}
	httpOptions := http.HttpOptions{
		Method:  "POST",
		Target:  target,
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

func TestHttpPathDnameParser(t *testing.T) {
	addresses := make(map[string][]string, 1)
	addresses["dname"] = []string{"soarca.tno.nl"}

	testTarget := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(80)}
	parsedUrl, err := http.ExtractUrl(testTarget, "/url")
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, parsedUrl, "http://soarca.tno.nl:80/url")
}

func TestHttpPathDnamePortParser(t *testing.T) {
	addresses := make(map[string][]string, 1)
	addresses["dname"] = []string{"soarca.tno.nl"}

	testTarget := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(8080)}
	parsedUrl, err := http.ExtractUrl(testTarget, "/url")
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, parsedUrl, "http://soarca.tno.nl:8080/url")
}

func TestHttpPathDnameRandomPortParser(t *testing.T) {
	addresses := make(map[string][]string, 1)
	addresses["dname"] = []string{"soarca.tno.nl"}

	testTarget := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(6464)}
	parsedUrl, err := http.ExtractUrl(testTarget, "/url")
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, parsedUrl, "https://soarca.tno.nl:6464/url")
}

func TestHttpPathIpv4Parser(t *testing.T) {
	addresses := make(map[string][]string, 1)
	addresses["ipv4"] = []string{"127.0.0.1"}

	testTarget := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(443)}
	parsedUrl, err := http.ExtractUrl(testTarget, "")
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, parsedUrl, "https://127.0.0.1:443")
}

func TestHttpPathParser(t *testing.T) {
	testTarget := cacao.AgentTarget{HttpUrl: "https://godcapability.tno.nl"}
	parsedUrl, err := http.ExtractUrl(testTarget, "")
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, parsedUrl, "https://godcapability.tno.nl")
}

func TestHttpPathBreakingParser(t *testing.T) {
	testTarget := cacao.AgentTarget{HttpUrl: "https://"}
	parsedUrl, err := http.ExtractUrl(testTarget, "")
	if err == nil {
		t.Error("want error for invalid empty domainname")
	}
	t.Logf(fmt.Sprint(parsedUrl))
}
