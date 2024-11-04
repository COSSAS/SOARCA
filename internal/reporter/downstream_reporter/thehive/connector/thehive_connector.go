package connector

import (
	"bytes"
	"encoding/json"
	"errors"
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

// ############################### ITheHiveConnector interface

type ITheHiveConnector interface {
	Hello() string
	PostStepTaskInCase(caseId string, step cacao.Step) error
	PostNewExecutionCase(executionId string, playbook cacao.Playbook) (string, error)
}

// ############################### TheHiveConnector object

type TheHiveConnector struct {
	baseUrl string
	apiKey  string
	ids_map SOARCATheHiveMap
}

func New(theHiveEndpoint string, theHiveApiKey string) *TheHiveConnector {
	ids_map := SOARCATheHiveMap{}
	ids_map.executionsCaseMaps = map[string]ExecutionCaseMap{}
	return &TheHiveConnector{
		baseUrl: theHiveEndpoint,
		apiKey:  theHiveApiKey,
		ids_map: ids_map,
	}
}

// ############################### Playbook to TheHive ID mappings

type SOARCATheHiveMap struct {
	executionsCaseMaps map[string]ExecutionCaseMap
}
type ExecutionCaseMap struct {
	caseId                  string
	stepsTasksMap           map[string]string
	variablesObservablesMap map[string]string
}

// TODO: Change to using observables instead of updating the tasks descriptions

func (soarcaTheHiveMap *SOARCATheHiveMap) checkExecutionCaseExists(executionId string) error {
	if _, ok := soarcaTheHiveMap.executionsCaseMaps[executionId]; !ok {
		return fmt.Errorf("case not found for execution id %s", executionId)
	}
	return nil
}
func (soarcaTheHiveMap *SOARCATheHiveMap) checkExecutionStepTaskExists(executionId string, stepId string) error {
	if _, ok := soarcaTheHiveMap.executionsCaseMaps[executionId].stepsTasksMap[stepId]; !ok {
		return fmt.Errorf("task not found for execution id %s for step id %s", executionId, stepId)
	}
	return nil
}

func (soarcaTheHiveMap *SOARCATheHiveMap) checkExecutionVariableObservableExists(executionId string, variableName string) error {
	if _, ok := soarcaTheHiveMap.executionsCaseMaps[executionId].variablesObservablesMap[variableName]; !ok {
		return fmt.Errorf("variable not found: execution id %s; variable: %s", executionId, variableName)
	}
	return nil
}

func (soarcaTheHiveMap *SOARCATheHiveMap) registerExecutionInCase(executionId string, caseId string) error {
	soarcaTheHiveMap.executionsCaseMaps[executionId] = ExecutionCaseMap{
		caseId:                  caseId,
		stepsTasksMap:           map[string]string{},
		variablesObservablesMap: map[string]string{},
	}
	//fmt.Printf("registering execution: %s, case id: %s", executionId, caseId)
	//fmt.Printf("execution entry id %s", soarcaTheHiveMap.executionsCaseMaps[executionId].caseId)
	log.Debugf("registering execution: %s, case id: %s", executionId, caseId)
	log.Debugf("execution entry id %s", soarcaTheHiveMap.executionsCaseMaps[executionId].caseId)

	return nil
}
func (soarcaTheHiveMap *SOARCATheHiveMap) registerStepTaskInCase(executionId string, stepId string, taskId string) error {
	soarcaTheHiveMap.executionsCaseMaps[executionId].stepsTasksMap[stepId] = taskId
	return nil
}

func (soarcaTheHiveMap *SOARCATheHiveMap) registerVariableObservableInCase(executionId string, variableName string, observableId string) error {
	soarcaTheHiveMap.executionsCaseMaps[executionId].variablesObservablesMap[variableName] = observableId
	fmt.Printf("registering observable: %s, variable name: %s", observableId, variableName)
	fmt.Printf("observable entry entry %s", soarcaTheHiveMap.executionsCaseMaps[executionId].variablesObservablesMap[variableName])
	return nil
}

func (soarcaTheHiveMap *SOARCATheHiveMap) retrieveCaseId(executionId string) (string, error) {
	err := soarcaTheHiveMap.checkExecutionCaseExists(executionId)
	if err != nil {
		return "", err
	}
	return soarcaTheHiveMap.executionsCaseMaps[executionId].caseId, nil
}

func (soarcaTheHiveMap *SOARCATheHiveMap) retrieveTaskId(executionId string, stepId string) (string, error) {
	err := soarcaTheHiveMap.checkExecutionCaseExists(executionId)
	if err != nil {
		return "", err
	}
	err = soarcaTheHiveMap.checkExecutionStepTaskExists(executionId, stepId)
	if err != nil {
		return "", err
	}
	return soarcaTheHiveMap.executionsCaseMaps[executionId].stepsTasksMap[stepId], nil
}

func (soarcaTheHiveMap *SOARCATheHiveMap) retrieveObservableId(executionId string, variableName string) (string, error) {
	err := soarcaTheHiveMap.checkExecutionCaseExists(executionId)
	if err != nil {
		return "", err
	}
	err = soarcaTheHiveMap.checkExecutionVariableObservableExists(executionId, variableName)
	if err != nil {
		return "", err
	}
	return soarcaTheHiveMap.executionsCaseMaps[executionId].variablesObservablesMap[variableName], nil
}

// func (soarcaTheHiveMap *SOARCATheHiveMap) clearCase(executionId string) error {
// 	err := soarcaTheHiveMap.checkExecutionCaseExists(executionId)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (soarcaTheHiveMap *SOARCATheHiveMap) clearMap(executionId string) error

// ############################### Functions

func (theHiveConnector *TheHiveConnector) PostStepTaskInCase(executionId string, step cacao.Step) error {
	caseId, err := theHiveConnector.ids_map.retrieveCaseId(executionId)
	if err != nil {
		return err
	}
	url := theHiveConnector.baseUrl + "/case/" + caseId + "/task"
	method := "POST"

	taskDescription := step.Description + "\n" + fmt.Sprintf("(SOARCA step: %s )", step.ID)
	task := schemas.Task{
		Title:       step.Name,
		Description: taskDescription,
	}

	body, err := theHiveConnector.sendRequest(method, url, task)
	if err != nil {
		return err
	}

	task_id, err := theHiveConnector.getIdFromRespBody(body)
	if err != nil {
		return err
	}
	theHiveConnector.ids_map.registerStepTaskInCase(executionId, step.ID, task_id)

	return nil
}

func (theHiveConnector *TheHiveConnector) PostVariableObservableInCasebyExecutionId(executionId string, variable cacao.Variable) error {
	caseId, err := theHiveConnector.ids_map.retrieveCaseId(executionId)
	if err != nil {
		return err
	}

	url := theHiveConnector.baseUrl + "/case/" + caseId + "/observable"
	method := "POST"

	observable := schemas.Observable{
		DataType: schemas.ObservableTypeOther,
		Data:     variable.Type + "\n" + variable.Name,
		Message:  variable.Description,
		TLP:      4,
		Tags:     []string{"CACAO Variable", executionId, variable.Type, variable.Value},
	}

	body, err := theHiveConnector.sendRequest(method, url, observable)
	if err != nil {
		return err
	}
	observableId, err := theHiveConnector.getIdFromRespBody(body)
	if err != nil {
		return err
	}

	theHiveConnector.ids_map.registerVariableObservableInCase(executionId, variable.Name, observableId)

	return nil
}

// TODO: revise this function through

func (theHiveConnector *TheHiveConnector) PostNewExecutionCase(executionId string, playbook cacao.Playbook) (string, error) {
	log.Tracef("posting new case to The Hive. execution ID %s, playbook %+v", executionId, playbook)

	url := theHiveConnector.baseUrl + "/case"
	method := "POST"

	// Add execution ID and playbook ID to tags (first and second tags)
	caseTags := []string{executionId, playbook.ID}
	caseTags = append(caseTags, playbook.Labels...)

	data := schemas.Case{
		Title:       playbook.Name,
		Description: playbook.Description,
		StartDate:   time.Now().Unix(),
		Tags:        caseTags,
	}

	body, err := theHiveConnector.sendRequest(method, url, data)
	if err != nil {
		return "", err
	}

	caseId, err := theHiveConnector.getIdFromRespBody(body)
	if err != nil {
		return "", err
	}

	err = theHiveConnector.ids_map.registerExecutionInCase(executionId, caseId)
	if err != nil {
		return "", err
	}

	// Pre-populate tasks according to playbook steps
	for _, step := range playbook.Workflow {
		err := theHiveConnector.PostStepTaskInCase(executionId, step)
		if err != nil {
			return "", err
		}
	}

	// Add variables as observables in case
	for _, variable := range playbook.PlaybookVariables {
		err := theHiveConnector.PostVariableObservableInCasebyExecutionId(executionId, variable)
		if err != nil {
			return "", err
		}
	}

	log.Tracef("case posted with case ID: %s", caseId)
	return string(body), nil
}

// TODO: finish and revise this function through

func (theHiveConnector *TheHiveConnector) UpdateEndExecutionCase(executionId string, variables cacao.Variables) (string, error) {
	caseId, err := theHiveConnector.ids_map.retrieveCaseId(executionId)
	if err != nil {
		return "", err
	}
	log.Tracef("updating case status to The Hive. execution ID %s, The Hive case ID %s", executionId, caseId)

	url := theHiveConnector.baseUrl + "/case/" + caseId
	method := "PATCH"

	data := schemas.Case{
		EndDate: time.Now().Unix(),
	}

	body, err := theHiveConnector.sendRequest(method, url, data)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// TODO: revise this function through

func (theHiveConnector *TheHiveConnector) UpdateStartStepTaskInCase(executionId string, step cacao.Step, variables cacao.Variables) (string, error) {
	log.Tracef("updating task in thehive. case ID %s", executionId)
	taskId, err := theHiveConnector.ids_map.retrieveTaskId(executionId, step.ID)
	if err != nil {
		return "", err
	}

	url := theHiveConnector.baseUrl + "/task/" + taskId + "/task"
	method := "PATCH"

	fullyAuto := true
	for _, command := range step.Commands {
		if command.Type == cacao.CommandTypeManual {
			fullyAuto = false
		}
	}
	taskAssignee := ""
	if fullyAuto {
		taskAssignee = "SOARCA Automated Execution"
	}
	task := schemas.Task{
		// StartDate: wait for merging of new reporting interface with timings passing,
		StartDate: time.Now().Unix(),
		Status:    schemas.TheHiveStatusInProgress,
		Assignee:  taskAssignee,
	}

	body, err := theHiveConnector.sendRequest(method, url, task)
	if err != nil {
		return "", err
	}

	return theHiveConnector.getIdFromRespBody(body)
}

func (theHiveConnector *TheHiveConnector) UpdateVariableObservableInCaseByExecutionId(executionId string, variable cacao.Variable) (string, error) {
	observableId, err := theHiveConnector.ids_map.retrieveObservableId(executionId, variable.Name)
	if err != nil {
		return "", err
	}

	url := theHiveConnector.baseUrl + "/observable/" + observableId
	method := "PATCH"

	observableUpdate := schemas.ObservableUpdate{
		AddTags: []string{variable.Value},
	}

	body, err := theHiveConnector.sendRequest(method, url, observableUpdate)
	if err != nil {
		return "", err
	}

	return theHiveConnector.getIdFromRespBody(body)
}

// TODO: revise this function through

func (theHiveConnector *TheHiveConnector) UpdateEndStepTaskInCase(executionId string, step cacao.Step, variables cacao.Variables, executionError error) (string, error) {
	taskId, err := theHiveConnector.ids_map.retrieveTaskId(executionId, step.ID)
	if err != nil {
		return "", err
	}

	url := theHiveConnector.baseUrl + "/task/" + taskId + "/task"
	method := "PATCH"

	taskStatus := schemas.TheHiveStatusCompleted
	if executionError != nil {
		taskStatus = "ExecutionError"
	}
	task := schemas.Task{
		// StartDate: wait for merging of new reporting interface with timings passing,
		EndDate: time.Now().Unix(),
		Status:  taskStatus,
	}

	body, err := theHiveConnector.sendRequest(method, url, task)
	if err != nil {
		return "", err
	}

	return theHiveConnector.getIdFromRespBody(body)
}

// ############################### HTTP interaction

func (theHiveConnector *TheHiveConnector) getIdFromRespBody(body []byte) (string, error) {
	var resp_map map[string]interface{}
	err := json.Unmarshal(body, &resp_map)
	if err != nil {
		return "", err
	}

	_id, ok := resp_map["_id"].(string)
	if !ok {
		// Handle the case where the type assertion fails
		log.Error("type assertion for retrieving TheHive ID failed")
		return "", errors.New("type assertion for retrieving TheHive ID failed")
	}

	return _id, nil
}

func (theHiveConnector *TheHiveConnector) sendRequest(method string, url string, body interface{}) ([]byte, error) {

	// Replace double slashes in the URL after http(s)://
	parts := strings.SplitN(url, "//", 2)
	cleanedPath := strings.ReplaceAll(parts[1], "//", "/")
	url = parts[0] + "//" + cleanedPath

	log.Tracef("sending request: %s %s", method, url)
	fmt.Printf("sending request: %s %s", method, url)

	var requestBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshalling JSON: %v", err)
		}
		fmt.Printf("Body: %s", jsonData)
		requestBody = bytes.NewBuffer(jsonData)
	}
	log.Debugf("request body: %s", requestBody)
	//fmt.Printf("request body: %s", requestBody)

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

// ############################### Utils
func PrettyJSONString(object interface{}) string {
	pretty, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		log.Warning("Error marshalling object to JSON:")
	}
	return string(pretty)
}
