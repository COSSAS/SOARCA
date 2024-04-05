package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"soarca/logger"
	"soarca/models/cacao"
)

var (
	component = reflect.TypeOf(HttpRequest{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type HttpOptions struct {
	Target  *cacao.AgentTarget
	Command *cacao.Command
	Auth    *cacao.AuthenticationInformation
}

type IHttpOptions interface {
	ExtractUrl() (string, error)
}
type IHttpRequest interface {
	Request(httpOptions HttpOptions) ([]byte, error)
}

type HttpRequest struct{}

// https://gist.githubusercontent.com/ahmetozer/ffa4cd0b319aff32ea9ed0068c8b81cf/raw/fc8742e6e087451e954bf0da214794a620356a4d/IPv4-IPv6-domain-regex.go
const (
	ipv6Regex   = `^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`
	ipv4Regex   = `^(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4})`
	domainRegex = `^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`
)

func (httpRequest *HttpRequest) Request(httpOptions HttpOptions) ([]byte, error) {
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

	httpOptions.addHeaderTo(request)
	err = httpOptions.addAuthTo(request)
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

func verifyAuthInfoMatchesAgentTarget(
	target *cacao.AgentTarget, authInfo *cacao.AuthenticationInformation,
) error {
	if target.AuthInfoIdentifier == "" || authInfo.ID == "" {
		return errors.New("target target.AuthInfoIndentifier or authInfo.ID is empty")
	}

	if !(target.AuthInfoIdentifier == authInfo.ID) {
		return errors.New("target auth info Id does not match auth info object's")
	}
	return nil
}

func (httpOptions *HttpOptions) addHeaderTo(request *http.Request) {
	for headerKey, headerValues := range httpOptions.Command.Headers {
		for _, headerValue := range headerValues {
			request.Header.Add(headerKey, headerValue)
		}
	}
}

func (httpOptions *HttpOptions) addAuthTo(request *http.Request) error {
	if httpOptions.Auth == nil {
		return nil
	}
	if (cacao.AuthenticationInformation{}) == *httpOptions.Auth {
		return nil
	}
	if err := verifyAuthInfoMatchesAgentTarget(httpOptions.Target, httpOptions.Auth); err != nil {
		return errors.New("auth info does not match target Id")
	}

	authInfoType := httpOptions.Auth.Type

	switch authInfoType {
	case cacao.AuthInfoHTTPBasicType:
		request.SetBasicAuth(httpOptions.Auth.Username, httpOptions.Auth.Password)
	case cacao.AuthInfoOAuth2Type:
		bearer := fmt.Sprintf("Bearer %s", httpOptions.Auth.Token)
		request.Header.Add("Authorization", bearer)
	case "":
		// It means that AuthN information is not set
		return nil
	default:
		return errors.New("unsupported authentication type: " + authInfoType)
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

	if len(target.Address) == 0 {
		return "", errors.New("cacao.AgentTarget does not contain enough information to build a proper query path")
	}

	if target.Port != "" {
		if err := validatePort(target.Port); err != nil {
			return "", err
		}
	}

	// If for an http-api command the agent-target address is a URL, it must be handled differently
	if len(target.Address["url"]) > 0 {
		if target.Address["url"][0] != "" {
			urlObject, err := parsePathBasedUrl(target.Address["url"][0])
			if err != nil {
				return "", err
			}
			if (urlObject.Path != "" && urlObject.Path != "/") && urlObject.Path != path {
				log.Warn("agent-target url has path that does not match http-api command path")
			}
			return urlObject.String(), nil
		}
	}
	return buildSchemeAndHostname(path, target)
}

func buildSchemeAndHostname(path string, target *cacao.AgentTarget) (string, error) {
	var hostname string

	scheme := setDefaultScheme(target)
	hostname, err := extractHostname(target)
	if err != nil {
		return "", err
	}

	parsedUrl := &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%s", hostname, target.Port),
		Path:   path,
	}

	return parsedUrl.String(), nil
}

func setDefaultScheme(target *cacao.AgentTarget) string {
	if target.Port == "" {
		target.Port = "80"
	}

	// Set the default scheme to HTTPS
	scheme := "https"
	if target.Port == "80" || target.Port == "8080" {
		scheme = "http"
	}
	return scheme
}

func extractHostname(target *cacao.AgentTarget) (string, error) {
	var address string

	if len(target.Address["dname"]) > 0 {
		match, _ := regexp.MatchString(domainRegex, target.Address["dname"][0])
		if !match {
			return "", errors.New("failed regex rule for domain name")
		}
		address = target.Address["dname"][0]

	} else if len(target.Address["ipv4"]) > 0 {
		match, _ := regexp.MatchString(ipv4Regex, target.Address["ipv4"][0])
		if !match {
			return "", errors.New("failed regex rule for domain name")
		}
		address = target.Address["ipv4"][0]

	} else {
		return "", errors.New("unsupported target address type")
	}
	return address, nil
}

func parsePathBasedUrl(httpUrl string) (*url.URL, error) {
	parsedUrl, err := url.ParseRequestURI(httpUrl)
	if err != nil {
		return parsedUrl, err
	}
	if parsedUrl.Hostname() == "" {
		return parsedUrl, errors.New("no domain name")
	}
	return parsedUrl, nil
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
