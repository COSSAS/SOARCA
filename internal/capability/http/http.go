package http

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"reflect"
	"soarca/logger"
	"soarca/models/cacao"
	"strings"

	"github.com/google/uuid"
)

// Receive HTTP API command data from decomposer/executer
// Validate HTTP API call
// Run HTTP API call
// Return response

type HttpCapability struct {
}

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// What to do if there is no agent or target?
// And maybe no auth info either?

func (httpCapability *HttpCapability) Execute(
	executionId uuid.UUID,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variables map[string]cacao.Variables) (map[string]cacao.Variables, error) {

	// Get request data and handle errors
	method, url, errmethod := ObtainHttpMethodAndUrlFromCommand(command)
	if errmethod != nil {
		log.Error(errmethod)
		return map[string]cacao.Variables{}, errmethod
	}
	content_data, errcontent := ObtainHttpRequestContentDataFromCommand(command)
	if errcontent != nil {
		log.Error(errcontent)
		return variables, errcontent
	}

	// Setup request
	request, err := http.NewRequest(method, url, bytes.NewBuffer(content_data))
	if err != nil {
		log.Error(err)
		return map[string]cacao.Variables{}, err
	}

	for key, httpCapability := range command.Headers {
		request.Header.Add(key, httpCapability)
	}
	if target.ID != "" {
		if err := verifyAuthInfoMatchesAgentTarget(&target, &authentication); err != nil {
			log.Error(err)
			return map[string]cacao.Variables{}, err
		}

		if err := setupAuthHeaders(request, &authentication); err != nil {
			log.Error(err)
			return map[string]cacao.Variables{}, err
		}
	}

	// Perform request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Error(err)
		return map[string]cacao.Variables{}, err
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
		return map[string]cacao.Variables{}, err
	}
	respString := string(responseBytes)
	sc := response.StatusCode
	if sc < 200 || sc > 299 {
		return map[string]cacao.Variables{}, errors.New(respString)
	}

	return map[string]cacao.Variables{
		"__soarca_http_result__": {Name: "result", Value: respString}}, nil

}

func ObtainHttpMethodAndUrlFromCommand(
	command cacao.Command) (string, string, error) {
	parts := strings.Split(command.Command, " ")
	if len(parts) != 2 {
		return "", "", errors.New("method or url missing from command")
	}
	method := parts[0]
	url := parts[1]
	return method, url, nil
}

func verifyAuthInfoMatchesAgentTarget(
	target *cacao.AgentTarget, authInfo *cacao.AuthenticationInformation) error {
	if !(target.AuthInfoIdentifier == authInfo.ID) {
		return errors.New("target auth info id does not match auth info object's")
	}
	return nil
}

func setupAuthHeaders(request *http.Request, authInfo *cacao.AuthenticationInformation) error {

	authInfoType := authInfo.Type
	switch authInfoType {
	case cacao.AuthInfoHTTPBasicType:
		request.SetBasicAuth(authInfo.Username, authInfo.Password)
	case cacao.AuthInfoOAuth2Type:
		// TODO: verify correctness
		// (https://datatracker.ietf.org/doc/html/rfc6750#section-2.1)
		bearer := "Bearer " + authInfo.Token
		request.Header.Add("Authorization", bearer)
	case "":
		// It means that AuthN information is not set
		return nil
	default:
		return errors.New("unsupported authentication type: " + authInfoType)
	}
	return nil
}

func ObtainHttpRequestContentDataFromCommand(
	command cacao.Command) ([]byte, error) {
	// Reads if either command or command_b64 are populated, and
	// Returns a byte slice from either
	content := command.Content
	content_b64 := command.ContentB64

	var nil_content []byte

	if content == "" && content_b64 == "" {
		return nil_content, nil
	}

	if content != "" && content_b64 != "" {
		log.Warn("both content and content_b64 are populated. using content.")
		return []byte(content), nil
	}

	if content != "" {
		return []byte(content), nil
	}

	return hex.DecodeString(content_b64)
}
