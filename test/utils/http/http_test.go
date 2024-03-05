package ssh_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	b64 "encoding/base64"
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
	command := cacao.Command{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Command: &command,
		Target:  &target,
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
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Command: &command,
		Target:  &target,
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
	command := cacao.Command{
		Type:    "http-api",
		Command: "PUT / HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Command: &command,
		Target:  &target,
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
	command := cacao.Command{
		Type:    "http-api",
		Command: "DELETE / HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Command: &command,
		Target:  &target,
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
	command := cacao.Command{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Command: &command,
		Target:  &target,
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
	command := cacao.Command{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Command: &command,
		Target:  &target,
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
	command := cacao.Command{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Command:  &command,
		Target:   &target,
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
	base64EncodedBody := b64.StdEncoding.EncodeToString(requestBody)
	if err != nil {
		fmt.Println("error Marshall JSON:", err)
	}
	if len(requestBody) == 0 {
		t.Fatalf("empty response")
	}
	target := cacao.AgentTarget{HttpUrl: "https://httpbin.org/anything"}
	command := cacao.Command{
		Type:       "http-api",
		Command:    "POST / HTTP/1.1",
		Headers:    map[string]string{"accept": "application/json"},
		ContentB64: base64EncodedBody,
	}
	httpOptions := http.HttpOptions{
		Command: &command,
		Target:  &target,
	}
	response, err := httpRequest.Request(httpOptions)

	t.Log(string(response))
	if err != nil {
		t.Fatalf(fmt.Sprint("http post request with body content has failed: ", err))
	}

	// specific format used by httpbin.org
	var httpBinReponse httpBinResponseBody
	err = json.Unmarshal(response, &httpBinReponse)
	fmt.Println(httpBinReponse)
	if err != nil {
		t.Fatalf(fmt.Sprint("Could not unmashall reponse token for bearer test,", err))
	}
	assert.Equal(t, httpBinReponse.Data, base64EncodedBody)
}

func TestHttpPathDnameParser(t *testing.T) {
	addresses := make(map[string][]string, 1)
	addresses["dname"] = []string{"soarca.tno.nl"}

	target := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(80)}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /url HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, parsedUrl, "http://soarca.tno.nl:80/url")
}

func TestHttpPathDnamePortParser(t *testing.T) {
	addresses := make(map[string][]string, 1)
	addresses["dname"] = []string{"soarca.tno.nl"}

	target := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(8080)}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /url HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, parsedUrl, "http://soarca.tno.nl:8080/url")
}

func TestHttpPathDnameRandomPortParser(t *testing.T) {
	addresses := make(map[string][]string, 1)
	addresses["dname"] = []string{"http://soarca.tno.nl"}

	target := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(6464)}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /url HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, parsedUrl, "https://soarca.tno.nl:6464/url")
}

func TestHttpPathIpv4Parser(t *testing.T) {
	addresses := make(map[string][]string, 1)
	addresses["ipv4"] = []string{"127.0.0.1"}

	target := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(443)}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, parsedUrl, "https://127.0.0.1:443/")
}

func TestHttpPathParser(t *testing.T) {
	target := cacao.AgentTarget{HttpUrl: "https://godcapability.tno.nl"}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, parsedUrl, "https://godcapability.tno.nl")
}

func TestHttpPathBreakingParser(t *testing.T) {
	target := cacao.AgentTarget{HttpUrl: "https://"}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	httpOptions := http.HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err == nil {
		t.Error("want error for invalid empty domainname")
	}
	t.Logf(fmt.Sprint(parsedUrl))
}

func TestMethodExtract(t *testing.T) {
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /api1/newObject HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	method, err := http.GetMethodFrom(&command)
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, method, "POST")
}

func TestPathExtract(t *testing.T) {
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /api1/newObject HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	path, err := http.GetPathFrom(&command)
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, path, "/api1/newObject")
}

func TestVersionExtract(t *testing.T) {
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /api1/newObject HTTP/1.1",
		Headers: map[string]string{"accept": "application/json"},
	}
	version, err := http.GetVersionFrom(&command)
	if err != nil {
		t.Fatalf(fmt.Sprint("failed test because: ", err))
	}
	assert.Equal(t, version, "HTTP/1.1")
}

func TestCommandFailedExtract(t *testing.T) {
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /api1/newObject",
		Headers: map[string]string{"accept": "application/json"},
	}
	version, err := http.GetVersionFrom(&command)
	if err == nil {
		t.Error("should give error as only 2 values are provided")
	}
	assert.Equal(t, version, "")
}
