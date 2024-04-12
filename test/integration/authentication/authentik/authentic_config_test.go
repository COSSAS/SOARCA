package authentik_test

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var _ OIDCProvider = &authentikServer{}

type authentikServer struct{}

func newAuthentik() OIDCProvider {
	return &authentikServer{}
}

func (i *authentikServer) AddClient(clientID, redirectURL string) string {
	return "secret"
}

func (i *authentikServer) URL() string {
	return "http://localhost:9000"
}

func (i *authentikServer) Close() {}

type ConsentResponseBody struct {
	Status string `json:"status"`
	Data   struct {
		RedirectURI string `json:"redirect_uri"`
	} `json:"data"`
}

func authentikLogin(client *http.Client, port string, initialResp *http.Response) (*http.Response, error) {
	// post credentials
	body, _ := json.Marshal(map[string]any{
		"username":       "tom",
		"password":       "tom",
		"keepMeLoggedIn": false,
	})
	_, err := client.Post("http://localhost:9000/api/firstfactor", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// get consent page
	resp, err := client.Get("http://localhost:" + port + "/")
	if err != nil {
		return nil, err
	}

	// post consent
	data, _ := json.Marshal(map[string]any{
		"client_id":     "apitest-" + port,
		"consent":       true,
		"consent_id":    resp.Request.URL.Query().Get("consent_id"),
		"pre_configure": false,
	})
	resp, err = client.Post("http://localhost:9000/api/oidc/consent", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	consentResponseBody := &ConsentResponseBody{}
	err = json.NewDecoder(resp.Body).Decode(consentResponseBody)
	if err != nil {
		return nil, err
	}

	// get homepage
	return client.Get(consentResponseBody.Data.RedirectURI)
}
