package connector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"soarca/internal/reporter/downstream_reporter/thehive/schemas"
	"soarca/logger"
	"soarca/models/cacao"
	"strings"
	"time"
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
	PostNewCase(caseId string, playbook cacao.Playbook) (string, error)
}

// The TheHive connector itself

type TheHiveConnector struct {
	baseUrl string
	apiKey  string
}

func New(theHiveEndpoint string, theHiveApiKey string) *TheHiveConnector {
	return &TheHiveConnector{baseUrl: theHiveEndpoint, apiKey: theHiveApiKey}
}

func (theHiveConnector *TheHiveConnector) sendRequest(method string, url string, body interface{}) ([]byte, error) {

	// Replace double slashes in the URL after http(s)://
	parts := strings.SplitN(url, "//", 2)
	cleanedPath := strings.ReplaceAll(parts[1], "//", "/")
	url = parts[0] + "//" + cleanedPath

	log.Tracef("sending request: %s %s", method, url)

	var requestBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshalling JSON: %v", err)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}
	log.Debugf("request body: %s", requestBody)
	fmt.Printf("request body: %s", requestBody)

	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+theHiveConnector.apiKey)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("received non-2xx status code: %d\nURL: %s", resp.StatusCode, url)
	}

	return respbody, nil
}

func (theHiveConnector *TheHiveConnector) Hello() string {

	url := theHiveConnector.baseUrl + "/user/current"

	body, err := theHiveConnector.sendRequest("GET", url, nil)
	if err != nil {
		return "error"
	}

	return (string(body))
}

func (theHiveConnector *TheHiveConnector) PostNewCase(caseId string, playbook cacao.Playbook) (string, error) {
	log.Tracef("posting new case to thehive. case ID %s, playbook %+v", caseId, playbook)

	url := theHiveConnector.baseUrl + "/case"
	var tasks []schemas.Task
	for _, step := range playbook.Workflow {
		task := schemas.Task{
			Title:       step.Name,
			Description: step.Description,
		}
		tasks = append(tasks, task)
	}

	// Add execution ID and playbook ID to tags (first and second tags)
	data := schemas.Case{
		Title:       playbook.Name,
		Description: playbook.Description,
		StartDate:   time.Now().Unix(),
		Tags:        playbook.Labels,
		Tasks:       tasks,
	}

	// data_bytes, err := json.Marshal(data)
	// if err != nil {
	// 	return "", err
	// }

	body, err := theHiveConnector.sendRequest("POST", url, data)
	if err != nil {
		return "", err
	}

	// TODO: cleanup
	var resp_map map[string]interface{}
	err = json.Unmarshal(body, &resp_map)
	if err != nil {
		return "", err
	}
	fmt.Println(resp_map)
	// Print the map
	pretty, err := json.MarshalIndent(resp_map, "", "    ")
	if err != nil {
		return "", err
	}
	fmt.Println(string(pretty))

	// Return the HTTP status code and the response body
	return string(body), nil
}
