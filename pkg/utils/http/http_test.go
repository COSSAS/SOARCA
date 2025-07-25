package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"soarca/pkg/models/cacao"

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
	httpRequest := HttpRequest{}

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://httpbin.org/get"},
		},
	}
	command := cacao.Command{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Command: &command,
		Target:  &target,
	}
	response, err := httpRequest.Request(httpOptions)
	t.Log(string(response))
	if err != nil {
		t.Error("http get request test has failed: ", err)
	}
	if len(response) == 0 {
		t.Error("empty response")
	}
	t.Log(string(response))
}

func TestHttpPostConnection(t *testing.T) {
	httpRequest := HttpRequest{}

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://httpbin.org/post"},
		},
	}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Command: &command,
		Target:  &target,
	}
	response, err := httpRequest.Request(httpOptions)
	t.Log(string(response))
	if err != nil {
		t.Error("http post request test has failed: ", err)
	}
	if len(response) == 0 {
		t.Error("empty response")
	}
	t.Log(string(response))
}

func TestHttpPutConnection(t *testing.T) {
	httpRequest := HttpRequest{}
	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://httpbin.org/put"},
		},
	}
	command := cacao.Command{
		Type:    "http-api",
		Command: "PUT / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Command: &command,
		Target:  &target,
	}
	response, err := httpRequest.Request(httpOptions)
	if err != nil {
		t.Error("http put request test has failed: ", err)
	}
	if len(response) == 0 {
		t.Error("empty response")
	}
	t.Log(string(response))
}

func TestHttpDeleteConnection(t *testing.T) {
	httpRequest := HttpRequest{}
	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://httpbin.org/delete"},
		},
	}
	command := cacao.Command{
		Type:    "http-api",
		Command: "DELETE / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Command: &command,
		Target:  &target,
	}
	response, err := httpRequest.Request(httpOptions)
	if err != nil {
		t.Error("http delete request test has failed: ", err)
	}
	if len(response) == 0 {
		t.Error("empty response")
	}
	t.Log(string(response))
}

// test status codes handling

func TestHttpStatus200(t *testing.T) {
	httpRequest := HttpRequest{}
	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://httpbin.org/status/200"},
		},
	}
	command := cacao.Command{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Command: &command,
		Target:  &target,
	}
	// error codes are handled internally by request, with throw error != in 200 range
	_, err := httpRequest.Request(httpOptions)
	t.Log(err)
	if err != nil {
		t.Error("http get request test has failed: ", err)
		return
	}
}

func TestHttpBearerToken(t *testing.T) {
	bearerToken := "test_token"
	httpRequest := HttpRequest{}

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://httpbin.org/bearer"},
		},
		AuthInfoIdentifier: "d0c7e6a0-f7fe-464e-9935-e6b3443f5b91",
	}
	auth := cacao.AuthenticationInformation{
		Type:  cacao.AuthInfoOAuth2Type,
		Token: bearerToken,
		ID:    "d0c7e6a0-f7fe-464e-9935-e6b3443f5b91",
	}
	command := cacao.Command{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Command: &command,
		Target:  &target,
		Auth:    &auth,
	}
	response, err := httpRequest.Request(httpOptions)
	if err != nil {
		t.Error("http request has failed: ", err, response)
	}
	if len(response) == 0 {
		t.Error("empty response")
	}
	t.Log(string(response))

	var responseToken bearerResponse
	err = json.Unmarshal(response, &responseToken)
	if err != nil {
		t.Error("could not unmashall reponsetoken for bearer test: ", err)
	}
	assert.Equal(t, responseToken.Authenticated, true)
	assert.Equal(t, responseToken.Token, bearerToken)
}

func TestHttpBasicAuth(t *testing.T) {
	user_id := "test"
	password := "password"
	url := fmt.Sprintf("https://httpbin.org/basic-auth/%s/%s", user_id, password)
	httpRequest := HttpRequest{}

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {url},
		},
		AuthInfoIdentifier: "d0c7e6a0-f7fe-464e-9935-e6b3443f5b91",
	}

	auth := cacao.AuthenticationInformation{
		Type:     cacao.AuthInfoHTTPBasicType,
		UserId:   user_id,
		Password: password,
		ID:       "d0c7e6a0-f7fe-464e-9935-e6b3443f5b91",
	}

	command := cacao.Command{
		Type:    "http-api",
		Command: "GET / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Command: &command,
		Target:  &target,
		Auth:    &auth,
	}
	response, err := httpRequest.Request(httpOptions)
	if err != nil {
		t.Error("http auth Request has failed: ", err)
	}
	if len(response) == 0 {
		t.Error("empty response")
	}
	t.Log(string(response))
	var authResponse basicAuthReponse
	err = json.Unmarshal(response, &authResponse)
	if err != nil {
		t.Error("could not unmashall reponse token for bearer test,", err)
	}

	assert.Equal(t, authResponse.Authenticated, true)
	assert.Equal(t, authResponse.User, user_id)
}

func TestHttpPostWithContentConnection(t *testing.T) {
	httpRequest := HttpRequest{}

	testJsonObj := testJson{Id: "28818819", User: "test", Description: "very interesting description"}
	requestBody, err := json.Marshal(testJsonObj)
	body := "some payload body"
	if err != nil {
		t.Error("error Marshall JSON: ", err)
	}
	if len(requestBody) == 0 {
		t.Error("empty response")
	}

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://httpbin.org/anything"},
		},
	}

	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
		Content: body,
	}
	httpOptions := HttpOptions{
		Command: &command,
		Target:  &target,
	}
	response, err := httpRequest.Request(httpOptions)

	t.Log(string(response))
	if err != nil {
		t.Error("http post request with body content has failed: ", err)
	}

	// specific format used by httpbin.org
	var httpBinReponse httpBinResponseBody
	err = json.Unmarshal(response, &httpBinReponse)
	fmt.Println(httpBinReponse)
	if err != nil {
		t.Error("Could not unmashall reponse token for bearer test,", err)
	}
	assert.Equal(t, httpBinReponse.Data, body)
}

func TestHttpPostWithBase64ContentConnection(t *testing.T) {
	httpRequest := HttpRequest{}

	testJsonObj := testJson{Id: "28818819", User: "test", Description: "very interesting description"}
	requestBody, err := json.Marshal(testJsonObj)
	base64EncodedBody := base64.StdEncoding.EncodeToString(requestBody)
	if err != nil {
		t.Error("error Marshall JSON: ", err)
	}
	if len(requestBody) == 0 {
		t.Error("empty response")
	}

	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://httpbin.org/anything"},
		},
	}

	command := cacao.Command{
		Type:       "http-api",
		Command:    "POST / HTTP/1.1",
		Headers:    map[string][]string{"accept": {"application/json"}},
		ContentB64: base64EncodedBody,
	}
	httpOptions := HttpOptions{
		Command: &command,
		Target:  &target,
	}
	response, err := httpRequest.Request(httpOptions)

	t.Log(string(response))
	if err != nil {
		t.Error("http post request with body content has failed: ", err)
	}

	// specific format used by httpbin.org
	var httpBinReponse httpBinResponseBody
	err = json.Unmarshal(response, &httpBinReponse)
	fmt.Println(httpBinReponse)
	if err != nil {
		t.Error("Could not unmashall reponse token for bearer test,", err)
	}
	decodedObject, err := base64.StdEncoding.DecodeString(base64EncodedBody)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, httpBinReponse.Data, string(decodedObject))
}

func TestHttpPathDnameParser(t *testing.T) {
	addresses := make(map[cacao.NetAddressType][]string, 1)
	addresses["dname"] = []string{"soarca.tno.nl"}

	target := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(80)}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /url HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		t.Error("failed test because: ", err)
	}
	assert.Equal(t, parsedUrl, "http://soarca.tno.nl:80/url")
}

func TestHttpPathDnamePortParser(t *testing.T) {
	addresses := make(map[cacao.NetAddressType][]string, 1)
	addresses["dname"] = []string{"soarca.tno.nl"}

	target := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(8080)}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /url HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		t.Error("failed test because: ", err)
	}
	assert.Equal(t, parsedUrl, "http://soarca.tno.nl:8080/url")
}

func TestHttpPathDnameRandomPortParser(t *testing.T) {
	addresses := make(map[cacao.NetAddressType][]string, 1)
	addresses["dname"] = []string{"soarca.tno.nl"}

	target := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(6464)}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /url HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		t.Error("failed test because: ", err)
	}
	assert.Equal(t, parsedUrl, "https://soarca.tno.nl:6464/url")
}

func TestHttpPathIpv4Parser(t *testing.T) {
	addresses := make(map[cacao.NetAddressType][]string, 1)
	addresses["ipv4"] = []string{"127.0.0.1"}

	target := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(443)}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		t.Error("failed test because: ", err)
	}
	assert.Equal(t, parsedUrl, "https://127.0.0.1:443/")
}

func TestHttpPathParser(t *testing.T) {
	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://godcapability.tno.nl"},
		},
	}

	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		t.Error("failed test because: ", err)
	}
	assert.Equal(t, parsedUrl, "https://godcapability.tno.nl")
}

func TestHttpPathUrlComposition(t *testing.T) {
	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://godcapability.tno.nl/isp"},
		},
	}

	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /isp/cst HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		t.Error("failed test because: ", err)
	}
	// Duplication of path values if present is INTENDED behaviour and
	// a warning will be issued
	assert.Equal(t, parsedUrl, "https://godcapability.tno.nl/isp/isp/cst")
}

func TestHttpPathBreakingParser(t *testing.T) {
	target := cacao.AgentTarget{
		Address: map[cacao.NetAddressType][]string{
			"url": {"https://"},
		},
	}

	command := cacao.Command{
		Type:    "http-api",
		Command: "POST / HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err == nil {
		t.Error("want error for invalid empty domainname")
	}
	t.Log(parsedUrl)
}

func TestMethodExtract(t *testing.T) {
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /api1/newObject HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	method, err := GetMethodFrom(&command)
	if err != nil {
		t.Error("failed test because: ", err)
	}
	assert.Equal(t, method, "POST")
}

func TestPathExtract(t *testing.T) {
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /api1/newObject HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	path, err := GetPathFrom(&command)
	if err != nil {
		t.Error("failed test because: ", err)
	}
	assert.Equal(t, path, "/api1/newObject")
}

func TestVersionExtract(t *testing.T) {
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /api1/newObject HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	version, err := GetVersionFrom(&command)
	if err != nil {
		t.Error("failed test because: ", err)
	}
	assert.Equal(t, version, "HTTP/1.1")
}

func TestCommandFailedExtract(t *testing.T) {
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /api1/newObject",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	version, err := GetVersionFrom(&command)
	if err == nil {
		t.Error("should give error as only 2 values are provided")
	}
	assert.Equal(t, version, "")
}

func TestDnameWithInvalidPathParser(t *testing.T) {
	addresses := make(map[cacao.NetAddressType][]string, 1)
	addresses["dname"] = []string{"soarca.tno.nl/this/path/shouldnt/be/used"}

	target := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(6464)}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /url HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err == nil {
		t.Error("should fail as invalid dname")
	}
	assert.Equal(t, parsedUrl, "")
}

func TestHttpPathIpv4WithRandomPort(t *testing.T) {
	addresses := make(map[cacao.NetAddressType][]string, 1)
	addresses["ipv4"] = []string{"127.0.0.1"}

	target := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(6464)}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /url HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		t.Error("failed test because: ", err)
	}
	assert.Equal(t, parsedUrl, "https://127.0.0.1:6464/url")
}

func TestInvalidDname(t *testing.T) {
	addresses := make(map[cacao.NetAddressType][]string, 1)
	addresses["dname"] = []string{"https://soarca.tno.nl"}

	target := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(6464)}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /url HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()

	if err == nil {
		t.Error("should fail as invalid dname")
	}
	assert.Equal(t, parsedUrl, "")
}

func TestInvalidIpv4(t *testing.T) {
	addresses := make(map[cacao.NetAddressType][]string, 1)
	addresses["ipv4"] = []string{"https://127.0.0.1"}

	target := cacao.AgentTarget{Address: addresses, Port: strconv.Itoa(6464)}
	command := cacao.Command{
		Type:    "http-api",
		Command: "POST /url HTTP/1.1",
		Headers: map[string][]string{"accept": {"application/json"}},
	}
	httpOptions := HttpOptions{
		Target:  &target,
		Command: &command,
	}

	parsedUrl, err := httpOptions.ExtractUrl()

	if err == nil {
		t.Error("should fail as invalid ipv4")
	}
	assert.Equal(t, parsedUrl, "")
}
