package thehive

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"soarca/internal/reporter/downstream_reporter/thehive/thehive_models"
	"soarca/internal/reporter/downstream_reporter/thehive/thehive_utils"
	"soarca/logger"
	"soarca/models/cacao"
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
	UpdateStartStepTaskInCase(executionId string, step cacao.Step, at time.Time) (string, error)
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

	message := thehive_models.TaskLog{Message: note}

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

func (theHiveConnector *TheHiveConnector) postStepDataAsCommentInTaskLog(executionId string, step cacao.Step, note string) error {

	message := note + "\n"
	stepData, err := thehive_utils.StructToMDJSON(step)
	if err != nil {
		return err
	}

	message = message + stepData

	err = theHiveConnector.postCommentInTaskLog(executionId, step, message)
	if err != nil {
		return err
	}

	return nil
}

func (theHiveConnector *TheHiveConnector) postStepVariablesAsCommentInTaskLog(executionId string, step cacao.Step, note string) error {
	variablesString := note + "\n"
	for _, variable := range step.StepVariables {
		variableJson, err := thehive_utils.StructToMDJSON(variable)
		if err != nil {
			return err
		}
		variablesString = variablesString + variableJson
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

	message := thehive_models.MessagePost{Message: note}

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
		variableJson, err := thehive_utils.StructToTxtJSON(variable)
		if err != nil {
			return err
		}
		variablesString = variablesString + variableJson
	}

	err := theHiveConnector.postCommentInCase(executionId, variablesString)
	if err != nil {
		return err
	}

	return nil
}

func (theHiveConnector *TheHiveConnector) postNewStepTaskInCase(executionId string, step cacao.Step) error {
	caseId, err := theHiveConnector.ids_map.retrieveCaseId(executionId)
	if err != nil {
		return err
	}
	url := theHiveConnector.baseUrl + "/case/" + caseId + "/task"
	method := "POST"

	taskDescription := step.Description + "\n" + fmt.Sprintf("(SOARCA step: %s )", step.ID)
	task := thehive_models.Task{
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
	log.Info("TheHive connector posting new execution in case")
	log.Trace(fmt.Sprintf("posting new case to The Hive. execution ID %s, playbook %+v", executionId, playbook))
	url := theHiveConnector.baseUrl + "/case"
	method := "POST"

	// Add execution ID and playbook ID to tags (first and second tags)
	caseTags := []string{executionId, playbook.ID}
	caseTags = append(caseTags, playbook.Labels...)

	data := thehive_models.Case{
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

	log.Info("Executing register execution in case")

	err = theHiveConnector.ids_map.registerExecutionInCase(executionId, caseId)
	if err != nil {
		return "", err
	}

	// Pre-populate tasks according to playbook steps
	for _, step := range playbook.Workflow {
		if step.Type == cacao.StepTypeStart || step.Type == cacao.StepTypeEnd {
			continue
		}
		err := theHiveConnector.postNewStepTaskInCase(executionId, step)
		if err != nil {
			return "", err
		}
	}

	executionStartMessage := fmt.Sprintf("START\nplaybook ID\n\t\t[ %s ]\nexecution ID\n\t\t[ %s ]\nstarted at\n\t\t[ %s ]", playbook.ID, executionId, at.String())
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

	caseStatus := thehive_models.TheHiveCaseStatusTruePositive
	closureComment := fmt.Sprintf("END\nexecution ID\n\t\t[ %s ]\nended at\n\t\t[ %s ]", executionId, at.String())
	if workflowErr != nil {
		caseStatus = thehive_models.TheHiveCaseStatusIndeterminate
		closureComment = closureComment + fmt.Sprintf("execution error: %s", workflowErr)
	}
	err = theHiveConnector.postCommentInCase(executionId, closureComment)
	if err != nil {
		log.Warningf("could not add task log: %s", err)
	}

	data := thehive_models.Case{
		Status:  caseStatus,
		Summary: "summary not implemented yet. Look at the tasks :)",
	}

	body, err := theHiveConnector.sendRequest(method, url, data)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// TODO: revise this function through

func (theHiveConnector *TheHiveConnector) UpdateStartStepTaskInCase(executionId string, step cacao.Step, at time.Time) (string, error) {
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
	task := thehive_models.Task{
		Status:   thehive_models.TheHiveStatusInProgress,
		Assignee: taskAssignee,
	}

	body, err := theHiveConnector.sendRequest(method, url, task)
	if err != nil {
		return "", err
	}

	executionStartMessage := fmt.Sprintf("START\nexecution ID\t\t[ %s ]\nstep ID\t\t[ %s ]\nstarted at\t\t[ %s ]", executionId, step.ID, at.String())
	err = theHiveConnector.postCommentInTaskLog(executionId, step, executionStartMessage)
	if err != nil {
		log.Warningf("could post message to task: %s", err)
	}

	err = theHiveConnector.postStepDataAsCommentInTaskLog(executionId, step, "step data")
	if err != nil {
		log.Warningf("could not report step data in step task log: %s", err)
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

	err = theHiveConnector.postStepVariablesAsCommentInTaskLog(executionId, step, "returned variables")
	if err != nil {
		log.Warningf("could not report variables in step task log: %s", err)
	}

	taskStatus := thehive_models.TheHiveStatusCompleted
	executionEndMessage := fmt.Sprintf("END\nexecution ID\t\t[ %s ]\nstep ID\t\t[ %s ]\nended at\t\t[ %s ]", executionId, step.ID, at.String())

	if stepErr != nil {
		taskStatus = thehive_models.TheHiveStatusCancelled
		executionEndMessage = executionEndMessage + fmt.Sprintf("\nexecution error: %s", stepErr)
	}
	err = theHiveConnector.postCommentInTaskLog(executionId, step, executionEndMessage)
	if err != nil {
		log.Warningf("could post message to task: %s", err)
	}

	task := thehive_models.Task{
		Status: taskStatus,
	}

	body, err := theHiveConnector.sendRequest(method, url, task)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// ############################### HTTP interaction

func (theHiveConnector *TheHiveConnector) sendRequest(method string, url string, body interface{}) ([]byte, error) {
	log.Info("Sending request")
	log.Trace(fmt.Sprintf("sending request: %s %s", method, url))

	req, err := theHiveConnector.prepareRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
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

func (theHiveConnector *TheHiveConnector) prepareRequest(method string, url string, body interface{}) (*http.Request, error) {
	url = thehive_utils.CleanUrlString(url)

	requestBody, err := thehive_utils.MarhsalRequestBody(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+theHiveConnector.apiKey)
	req.Header.Add("Content-Type", "application/json")

	return req, nil
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

	id, err := thehive_utils.GetIdFromArrayBody(body)
	if err == nil {
		return id, err
	}

	id, err = thehive_utils.GetIdFromObjectBody(body)
	if err == nil {
		return id, err
	}

	log.Debug(fmt.Sprintf("body: %s", string(body)))
	return "", errors.New("failed to get ID from response body")

}
