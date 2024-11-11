package thehive

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"soarca/logger"
	"soarca/models/cacao"
	"strings"
	"time"
)

// TODOs
// Fix asynchronous http api calls causing The Hive reporting to be all over the place

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
	PostNewExecutionCase(executionId string, playbook cacao.Playbook, at time.Time) (string, error)
	UpdateEndExecutionCase(executionId string, variables cacao.Variables, workflowErr error, at time.Time) (string, error)
	UpdateStartStepTaskInCase(executionId string, step cacao.Step, variables cacao.Variables, at time.Time) (string, error)
	UpdateEndStepTaskInCase(executionId string, step cacao.Step, returnVars cacao.Variables, stepErr error, at time.Time) (string, error)
}

// ############################### TheHiveConnector object

type TheHiveConnector struct {
	baseUrl string
	apiKey  string
	ids_map *SOARCATheHiveMap
}

func NewConnector(theHiveEndpoint string, theHiveApiKey string) *TheHiveConnector {
	ids_map := &SOARCATheHiveMap{}
	ids_map.executionsCaseMaps = map[string]ExecutionCaseMap{}
	return &TheHiveConnector{
		baseUrl: theHiveEndpoint,
		apiKey:  theHiveApiKey,
		ids_map: ids_map,
	}
}

// ############################### Functions

func (theHiveConnector *TheHiveConnector) postCommentInTaskLog(executionId string, step cacao.Step, note string) error {
	log.Trace(fmt.Sprintf("posting comment in task log via execution ID: %s. step ID: %s", executionId, step.ID))
	taskId, err := theHiveConnector.ids_map.retrieveTaskId(executionId, step.ID)
	if err != nil {
		return err
	}
	url := theHiveConnector.baseUrl + "/task/" + taskId + "/log"
	method := "POST"

	message := TaskLog{Message: note}

	body, err := theHiveConnector.sendRequest(method, url, message)
	if err != nil {
		return err
	}
	messageId, err := theHiveConnector.getIdFromRespBody(body)
	if err != nil {
		return err
	}
	log.Trace(fmt.Sprintf("task log created. execution ID %s, Task id %s, message Id: %s", executionId, taskId, messageId))

	return nil
}

func (theHiveConnector *TheHiveConnector) postStepVariablesAsCommentInTaskLog(executionId string, step cacao.Step, note string) error {
	variablesString := note + "\n"
	for _, variable := range step.StepVariables {
		variablesString = variablesString + formatVariable(variable)
	}

	err := theHiveConnector.postCommentInTaskLog(executionId, step, variablesString)
	if err != nil {
		return err
	}

	return nil
}

func (theHiveConnector *TheHiveConnector) postCommentInCase(executionId string, note string) error {
	log.Info(fmt.Sprintf("posting comment in case via execution ID: %s.", executionId))
	caseId, err := theHiveConnector.ids_map.retrieveCaseId(executionId)
	if err != nil {
		return err
	}

	url := theHiveConnector.baseUrl + "/case/" + caseId + "/comment"
	method := "POST"

	message := MessagePost{Message: note}

	body, err := theHiveConnector.sendRequest(method, url, message)
	if err != nil {
		return err
	}
	messageId, err := theHiveConnector.getIdFromRespBody(body)
	if err != nil {
		return err
	}
	log.Trace(fmt.Sprintf("Case comment created. execution ID %s, caseId %s, message Id: %s", executionId, caseId, messageId))
	return nil
}

func (theHiveConnector *TheHiveConnector) postVariablesAsCommentInCase(executionId string, variables cacao.Variables, note string) error {

	variablesString := note + "\n"
	for _, variable := range variables {
		variablesString = variablesString + formatVariable(variable)
	}

	err := theHiveConnector.postCommentInCase(executionId, variablesString)
	if err != nil {
		return err
	}

	return nil
}

func (theHiveConnector *TheHiveConnector) registerStepTaskInCase(executionId string, step cacao.Step) error {
	caseId, err := theHiveConnector.ids_map.retrieveCaseId(executionId)
	if err != nil {
		return err
	}
	url := theHiveConnector.baseUrl + "/case/" + caseId + "/task"
	method := "POST"

	taskDescription := step.Description + "\n" + fmt.Sprintf("(SOARCA step: %s )", step.ID)
	task := Task{
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

// ######################################## Connector interface

func (theHiveConnector *TheHiveConnector) PostNewExecutionCase(executionId string, playbook cacao.Playbook, at time.Time) (string, error) {
	log.Trace(fmt.Sprintf("posting new case to The Hive. execution ID %s, playbook %+v", executionId, playbook))
	url := theHiveConnector.baseUrl + "/case"
	method := "POST"

	// Add execution ID and playbook ID to tags (first and second tags)
	caseTags := []string{executionId, playbook.ID}
	caseTags = append(caseTags, playbook.Labels...)

	data := Case{
		Title:       playbook.Name,
		Description: playbook.Description,
		//StartDate:   int(time.Now().Unix()),
		Tags: caseTags,
	}

	log.Tracef("sending request: %s %s", method, url)

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
		if strings.Contains(step.ID, "start") || strings.Contains(step.ID, "end") {
			continue
		}
		err := theHiveConnector.registerStepTaskInCase(executionId, step)
		if err != nil {
			return "", err
		}
	}

	executionStartMessage := fmt.Sprintf("START\nplaybook ID [ %s ]\nexecution ID [ %s ]\nstarted in SOARCA at: [ %s ]", playbook.ID, executionId, at.String())
	err = theHiveConnector.postCommentInCase(executionId, executionStartMessage)
	if err != nil {
		log.Warningf("could not post message to case: %s", err)
	}

	err = theHiveConnector.postVariablesAsCommentInCase(executionId, playbook.PlaybookVariables, "variables at start of execution")
	if err != nil {
		log.Warningf("could not report variables in case comment: %s", err)
	}

	log.Tracef("case posted with case ID: %s", caseId)
	return string(body), nil
}

func (theHiveConnector *TheHiveConnector) UpdateEndExecutionCase(executionId string, variables cacao.Variables, workflowErr error, at time.Time) (string, error) {
	caseId, err := theHiveConnector.ids_map.retrieveCaseId(executionId)
	if err != nil {
		return "", err
	}
	log.Trace(fmt.Sprintf("updating case status to The Hive. execution ID %s, The Hive case ID %s", executionId, caseId))

	url := theHiveConnector.baseUrl + "/case/" + caseId
	method := "PATCH"

	err = theHiveConnector.postVariablesAsCommentInCase(executionId, variables, "variables at end of execution")
	if err != nil {
		log.Warningf("could not add task log: %s", err)
	}

	caseStatus := TheHiveCaseStatusTruePositive
	closureComment := fmt.Sprintf("END\nexecution ID [ %s ]\nended in SOARCA at: [ %s ]", executionId, at.String())
	if workflowErr != nil {
		caseStatus = TheHiveCaseStatusIndeterminate
		closureComment = closureComment + fmt.Sprintf("execution error: %s", workflowErr)
	}
	err = theHiveConnector.postCommentInCase(executionId, closureComment)
	if err != nil {
		log.Warningf("could not add task log: %s", err)
	}

	data := Case{
		//EndDate: int(time.Now().Unix()),
		Status: caseStatus,
		//ImpactStatus: TheHiveCaseImpacStatustLow,
		Summary: "summary not implemented yet. Look at the tasks :)",
	}

	body, err := theHiveConnector.sendRequest(method, url, data)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// TODO: revise this function through

func (theHiveConnector *TheHiveConnector) UpdateStartStepTaskInCase(executionId string, step cacao.Step, variables cacao.Variables, at time.Time) (string, error) {
	log.Trace(fmt.Sprintf("updating task in thehive. case ID %s. task started.", executionId))
	taskId, err := theHiveConnector.ids_map.retrieveTaskId(executionId, step.ID)
	if err != nil {
		return "", err
	}

	url := theHiveConnector.baseUrl + "/task/" + taskId
	method := "PATCH"

	fullyAuto := true
	for _, command := range step.Commands {
		if command.Type == cacao.CommandTypeManual {
			fullyAuto = false
		}
	}

	// Must identify valid user in The hive. Cannot be custom string
	taskAssignee := "soarca@soarca.eu"
	if fullyAuto {
		taskAssignee = "soarca@soarca.eu"
	}
	task := Task{
		// StartDate: wait for merging of new reporting interface with timings passing,
		StartDate: time.Now().Unix(),
		Status:    TheHiveStatusInProgress,
		Assignee:  taskAssignee,
	}

	body, err := theHiveConnector.sendRequest(method, url, task)
	if err != nil {
		return "", err
	}

	executionStartMessage := fmt.Sprintf("START\nexecution ID [ %s ]\nstep ID [ %s ]\nstarted in SOARCA at: [ %s ]", executionId, step.ID, at.String())
	err = theHiveConnector.postCommentInTaskLog(executionId, step, executionStartMessage)
	if err != nil {
		log.Warningf("could post message to task: %s", err)
	}

	err = theHiveConnector.postStepVariablesAsCommentInTaskLog(executionId, step, "variables at start of step execution")
	if err != nil {
		log.Warningf("could not report variables in step task log: %s", err)
	}

	return theHiveConnector.getIdFromRespBody(body)
}

func (theHiveConnector *TheHiveConnector) UpdateEndStepTaskInCase(executionId string, step cacao.Step, returnVars cacao.Variables, stepErr error, at time.Time) (string, error) {
	log.Trace(fmt.Sprintf("updating task in thehive. case ID %s. task ended.", executionId))
	taskId, err := theHiveConnector.ids_map.retrieveTaskId(executionId, step.ID)
	if err != nil {
		return "", err
	}

	url := theHiveConnector.baseUrl + "/task/" + taskId
	method := "PATCH"

	err = theHiveConnector.postStepVariablesAsCommentInTaskLog(executionId, step, "variables at end of step execution")
	if err != nil {
		log.Warningf("could not report variables in step task log: %s", err)
	}

	taskStatus := TheHiveStatusCompleted
	executionEndMessage := fmt.Sprintf("END\nexecution ID [ %s ]\nstep ID [ %s ]\nended in SOARCA at: [ %s ]", executionId, step.ID, at.String())

	if stepErr != nil {
		taskStatus = TheHiveStatusCancelled
		executionEndMessage = executionEndMessage + fmt.Sprintf("\nexecution error: %s", stepErr)
	}
	err = theHiveConnector.postCommentInTaskLog(executionId, step, executionEndMessage)
	if err != nil {
		log.Warningf("could post message to task: %s", err)
	}

	task := Task{
		// StartDate: wait for merging of new reporting interface with timings passing,
		EndDate: time.Now().Unix(),
		Status:  taskStatus,
	}

	body, err := theHiveConnector.sendRequest(method, url, task)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// ############################### HTTP interaction

func (theHiveConnector *TheHiveConnector) sendRequest(method string, url string, body interface{}) ([]byte, error) {
	log.Trace(fmt.Sprintf("sending request: %s %s", method, url))

	// Replace double slashes in the URL after http(s)://
	parts := strings.SplitN(url, "//", 2)
	cleanedPath := strings.ReplaceAll(parts[1], "//", "/")
	url = parts[0] + "//" + cleanedPath

	var requestBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshalling JSON: %v", err)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}
	log.Debug(fmt.Sprintf("request body: %s", requestBody))

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

	log.Debug(fmt.Sprintf("response body: %s", respbody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("received non-2xx status code: %d\nURL: %s: %s", resp.StatusCode, url, respbody)
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

func (theHiveConnector *TheHiveConnector) getIdFromRespBody(body []byte) (string, error) {

	if len(body) == 0 {
		return "", nil
	}

	id, err := getIdFromArrayBody(body)
	if err == nil {
		return id, err
	}

	id, err = getIdFromObjectBody(body)
	if err == nil {
		return id, err
	}

	log.Debug(fmt.Sprintf("body: %s", string(body)))
	return "", errors.New("failed to get ID from response body")

}

// ############################### Utils
func PrettyPrintObject(object interface{}) string {
	pretty, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		log.Warningf("Error marshalling object to JSON: %s", err)
		return ""
	}
	return string(pretty) + "\n"
}

func getIdFromArrayBody(body []byte) (string, error) {
	// Try to unmarshal as a slice of maps
	var respArray []map[string]interface{}
	err := json.Unmarshal(body, &respArray)
	if err != nil {
		return "", err
	}

	if len(respArray) == 0 {
		return "", errors.New("response array is empty")
	}

	_id, ok := respArray[0]["_id"].(string)
	if !ok {
		log.Error("type assertion for retrieving TheHive ID failed")
		return "", errors.New("type assertion for retrieving TheHive ID failed")
	}
	return _id, nil
}

func getIdFromObjectBody(body []byte) (string, error) {
	// If unmarshalling as a slice fails, try to unmarshal as a single map
	var respMap map[string]interface{}
	err := json.Unmarshal(body, &respMap)
	if err != nil {
		return "", err
	}

	_id, ok := respMap["_id"].(string)
	if !ok {
		log.Error("type assertion for retrieving TheHive ID from map failed")
		return "", errors.New("type assertion for retrieving TheHive ID from map failed")
	}

	return _id, nil
}

func formatVariable(s interface{}) string {
	v := reflect.ValueOf(s)
	t := v.Type()

	var sb strings.Builder

	sb.WriteString(strings.Repeat("*", 30) + "\n")

	maxFieldLength := 0
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i).Name
		if len(field) > maxFieldLength {
			maxFieldLength = len(field)
		}
	}

	var name, value, description string
	var otherFields []string

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i).Name
		fieldValue := fmt.Sprintf("%v", v.Field(i).Interface())
		padding := strings.Repeat("\t", (maxFieldLength-len(field))/4+1)
		switch field {
		case "Name":
			name = fmt.Sprintf("%s:%s%s\n", field, padding, fieldValue)
		case "Value":
			value = fmt.Sprintf("%s:%s%s\n", field, padding, fieldValue)
		case "Description":
			description = fmt.Sprintf("%s:%s%s\n", field, padding, fieldValue)
		default:
			otherFields = append(otherFields, fmt.Sprintf("%s:%s%s\n", field, padding, fieldValue))
		}
	}

	// Append "Name", "Description", and "Value" fields first
	sb.WriteString(name)
	sb.WriteString(description)
	sb.WriteString(value)

	// Append all other fields
	for _, field := range otherFields {
		sb.WriteString(field)
	}

	sb.WriteString(strings.Repeat("*", 30) + "\n")

	return sb.String()
}
