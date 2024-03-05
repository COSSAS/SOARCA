package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"soarca/logger"
	"soarca/models/cacao"
)

var (
	component = reflect.TypeOf(HttpRequest{}).PkgPath()
	log       *logger.Log
)

type HttpOptions struct {
	Target   *cacao.AgentTarget
	Command  *cacao.Command
	Username string
	Password string
	Token    string
}

type IHttpRequest interface {
	Request() ([]byte, error)
}

type HttpRequest struct{}

func (httpRequest *HttpRequest) Request(httpOptions HttpOptions) ([]byte, error) {
	log = logger.Logger(component, logger.Info, "", logger.Json)
	request, err := httpOptions.setupRequest()
	if err != nil {
		return []byte{}, err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Error(err)
		return []byte{}, err
	}

	defer response.Body.Close()
	return httpOptions.handleResponse(response)
}

func (httpOptions *HttpOptions) setupRequest() (*http.Request, error) {
	parsedUrl, err := httpOptions.ExtractUrl()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	method, err := GetMethodFrom(httpOptions.Command)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	requestBuffer := bytes.NewBufferString(httpOptions.Command.ContentB64)
	request, err := http.NewRequest(method, parsedUrl, requestBuffer)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	err = httpOptions.populateRequestFields(request)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return request, nil
}

func (httpRequest *HttpOptions) handleResponse(response *http.Response) ([]byte, error) {
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
		return []byte{}, err
	}
	sc := response.StatusCode
	if sc < 200 || sc > 299 {
		return []byte{}, errors.New(string(responseBytes))
	}
	return responseBytes, nil
}

func (httpOptions *HttpOptions) populateRequestFields(request *http.Request) error {
	for header_key, header_value := range httpOptions.Command.Headers {
		request.Header.Add(header_key, header_value)
	}
	if (httpOptions.Username != "" && httpOptions.Token != "") || (httpOptions.Password != "" && httpOptions.Token != "") {
		return errors.New("both credentials and token are used in HTTP Request. Credentials are used, token is ignored")
	}
	if httpOptions.Username != "" || httpOptions.Password != "" {
		request.SetBasicAuth(httpOptions.Username, httpOptions.Password)
	}
	if httpOptions.Token != "" {
		bearer := "Bearer " + httpOptions.Token
		request.Header.Add("Authorization", bearer)
	}
	return nil
}

func (httpOptions *HttpOptions) ExtractUrl() (string, error) {
	if httpOptions.Command == nil || httpOptions.Target == nil {
		return "", errors.New("not enough http options supplied, nil found")
	}
	path, err := GetPathFrom(httpOptions.Command)
	if err != nil {
		log.Error(err)
		return "", err
	}
	target := httpOptions.Target

	if len(target.Address) == 0 && target.HttpUrl == "" {
		return "", errors.New("cacao.AgentTarget does not contain enough information to build a proper query path")
	}

	if target.Port != "" {
		if err := validatePort(target.Port); err != nil {
			return "", err
		}
	}

	if target.HttpUrl != "" {
		parsedUrl, err := url.ParseRequestURI(target.HttpUrl)
		if err != nil {
			return "", err
		}
		if parsedUrl.Host == "" {
			return "", errors.New("no domain name")
		}
		return parsedUrl.String(), nil
	}
	var hostname string

	// according to the cacao spec!
	if target.Port == "" {
		target.Port = "80"
	}

	// Set the default scheme to HTTPS
	scheme := "https"
	if target.Port == "80" || target.Port == "8080" {
		scheme = "http"
	}

	if len(target.Address["dname"]) > 0 {
		hostname = target.Address["dname"][0]
	} else if len(target.Address["ipv4"]) > 0 {
		hostname = target.Address["ipv4"][0]
	} else {
		return "", errors.New("unsupported target address type")
	}
	if hostname == "" {
		return "", errors.New("hostname or path remains empty")
	}

	parsedUrl := &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%s", hostname, target.Port),
		Path:   path,
	}

	return parsedUrl.String(), nil
}

func validatePort(port string) error {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return errors.New("could not parse string to port number")
	}
	if portNum < 1 || portNum > 65535 {
		return errors.New("port must be in the range 1-65535")
	}
	return nil
}

func GetMethodFrom(command *cacao.Command) (string, error) {
	return extractCommandFieldByIndex(command, 0)
}

func GetPathFrom(command *cacao.Command) (string, error) {
	return extractCommandFieldByIndex(command, 1)
}

func GetVersionFrom(command *cacao.Command) (string, error) {
	return extractCommandFieldByIndex(command, 2)
}

func extractCommandFieldByIndex(command *cacao.Command, index int) (string, error) {
	if command == nil {
		return "", errors.New("command pointer is empty")
	}
	if index < 0 || index > 2 {
		return "", errors.New("invalid index")
	}
	parts := strings.Fields(command.Command)
	if len(parts) != 3 {
		return "", errors.New("invalid request format")
	}
	if parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", errors.New("method, path, or HTTP version is empty")
	}
	return parts[index], nil
}
