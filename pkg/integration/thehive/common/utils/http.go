package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Sets up a client with TLS config
func SetupClient(allowInsecure bool) *http.Client {

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: allowInsecure},
	}

	client := &http.Client{Transport: transport}
	return client

}

func SendRequest(client *http.Client, req *http.Request) ([]byte, error) {
	if client == nil || req == nil {
		return nil, errors.New("payload is nil")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	log.Debug(fmt.Sprintf("response body: %s", respbody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("received non-2xx status code: %d\nURL: %s: %s", resp.StatusCode, req.Host, respbody)
	}

	return respbody, nil
}

func PrepareRequest(method string, url string, apiKey string, body interface{}) (*http.Request, error) {
	log.Trace(fmt.Sprintf("sending request: %s %s", method, url))

	requestBody, err := MarhsalRequestBody(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}
