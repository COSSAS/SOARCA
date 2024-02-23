package http

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"reflect"

	"soarca/logger"
)

var (
	component = reflect.TypeOf(HttpRequest{}).PkgPath()
	log       *logger.Log
)

type HttpOptions struct {
	Method   string
	Url      string
	Body     []byte
	Headers  map[string]string
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
	request, err := httpOptions.setup()
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

func (httpRequest *HttpOptions) setup() (*http.Request, error) {
	request, err := http.NewRequest(httpRequest.Method, httpRequest.Url, bytes.NewBuffer(httpRequest.Body))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	err = httpRequest.populateRequestFields(request)
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
	for header_key, header_value := range httpOptions.Headers {
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
