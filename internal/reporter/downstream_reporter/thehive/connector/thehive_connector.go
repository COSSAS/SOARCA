package connector

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"soarca/logger"
	"soarca/models/cacao"
)

var (
	component = reflect.TypeOf(TheHiveConnector{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

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

func (theHiveConnector *TheHiveConnector) PostNewCase(caseId string, playbook cacao.Playbook) error {

	url := theHiveConnector.baseUrl + "/user/current"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	token := theHiveConnector.apiKey
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Print the response body
	fmt.Println(string(body))
	return err
}
