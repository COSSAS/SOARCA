package connector

import (
	"fmt"
	"io"
	"net/http"
)

type ITheHiveConnector interface {
	Hello() string
}

// The TheHive connector itself

type TheHiveConnector struct {
	baseUrl string
	apiKey  string
}

func New(theHiveEndpoint string, theHiveApiKey string) *TheHiveConnector {
	return &TheHiveConnector{baseUrl: theHiveEndpoint, apiKey: theHiveApiKey}
}

func (theHiveConnector *TheHiveConnector) Hello() string {

	url := theHiveConnector.baseUrl + "/user/current"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	token := theHiveConnector.apiKey
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	// Print the response body
	fmt.Println(string(body))
	return ""
}
