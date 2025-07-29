package connector

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/integration/thehive/common/mappings"
	thehive_models "soarca/pkg/integration/thehive/common/models"
	thehive_utils "soarca/pkg/integration/thehive/common/utils"
	"soarca/pkg/models/cacao"
	"time"
)

const theHiveV1ApiPath = "thehive/api/v1"
const theHiveCasePath = "/case"
const theHiveObservablePath = "/observable"

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
	PostNewExecutionCase(executionMetadata thehive_models.ExecutionMetadata, at time.Time) (string, error)
	UpdateEndExecutionCase(executionMetadata thehive_models.ExecutionMetadata, at time.Time) (string, error)
	UpdateStartStepTaskInCase(executionMetadata thehive_models.ExecutionMetadata, at time.Time) (string, error)
	UpdateEndStepTaskInCase(executionMetadata thehive_models.ExecutionMetadata, at time.Time) (string, error)
}

// ############################### TheHiveConnector object

type TheHiveConnector struct {
	baseUrl string
	apiKey  string
	ids_map *mappings.SOARCATheHiveMap
	client  *http.Client
}

func NewConnector(theHiveEndpoint string, theHiveApiKey string, allowInsecure bool) *TheHiveConnector {
	ids_map := &mappings.SOARCATheHiveMap{}
	ids_map.ExecutionsCaseMaps = map[string]mappings.ExecutionCaseMap{}
	return &TheHiveConnector{
		client:  thehive_utils.SetupClient(allowInsecure),
		baseUrl: theHiveEndpoint,
		apiKey:  theHiveApiKey,
		ids_map: ids_map,
	}

}

// ############################### Functions

func (theHiveConnector *TheHiveConnector) CreateCase(thisCase thehive_models.Case) error {
	path := theHiveConnector.baseUrl + theHiveCasePath

	method := "POST"

	request, err := thehive_utils.PrepareRequest(method, path, theHiveConnector.apiKey, thisCase)
	if err != nil {
		log.Error(err)
		return err
	}
	thehive_utils.SendRequest(theHiveConnector.client, request)
	return nil
}

///

func (theHiveConnector *TheHiveConnector) postCommentInTaskLog(executionId string, step cacao.Step, note string) error {
	log.Trace(fmt.Sprintf("posting comment in task log via execution ID: %s. step ID: %s", executionId, step.ID))
	taskId, err := theHiveConnector.ids_map.RetrieveTaskId(executionId, step.ID)
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
	log.Trace(fmt.Sprintf("posting comment in case via execution ID: %s.", executionId))
	caseId, err := theHiveConnector.ids_map.RetrieveCaseId(executionId)
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
	caseId, err := theHiveConnector.ids_map.RetrieveCaseId(executionId)
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
	theHiveConnector.ids_map.RegisterStepTaskInCase(executionId, step.ID, task_id)

	return nil
}

// ######################################## Connector interface

func (theHiveConnector *TheHiveConnector) PostNewExecutionCase(execMetadata thehive_models.ExecutionMetadata, at time.Time) (string, error) {
	log.Trace(fmt.Sprintf("posting new case to The Hive. execution ID %s, playbook %+v", execMetadata.ExecutionId, execMetadata.Playbook))
	url := theHiveConnector.baseUrl + "/case"
	method := "POST"

	// Add execution ID and playbook ID to tags (first and second tags)
	caseTags := []string{execMetadata.ExecutionId, execMetadata.Playbook.ID}
	caseTags = append(caseTags, execMetadata.Playbook.Labels...)

	data := thehive_models.Case{
		Title:       execMetadata.Playbook.Name,
		Description: execMetadata.Playbook.Description,
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

	log.Trace("Executing register execution in case")

	err = theHiveConnector.ids_map.RegisterExecutionInCase(execMetadata.ExecutionId, caseId)
	if err != nil {
		return "", err
	}

	// Pre-populate tasks according to playbook steps
	for _, step := range execMetadata.Playbook.Workflow {
		if step.Type == cacao.StepTypeStart || step.Type == cacao.StepTypeEnd {
			continue
		}
		err := theHiveConnector.postNewStepTaskInCase(execMetadata.ExecutionId, step)
		if err != nil {
			return "", err
		}
	}

	executionStartMessage := fmt.Sprintf(
		"START\nplaybook ID\n\t\t[ %s ]\nexecution ID\n\t\t[ %s ]\nstarted at\n\t\t[ %s ]",
		execMetadata.Playbook.ID, execMetadata.ExecutionId, at.String())
	err = theHiveConnector.postCommentInCase(execMetadata.ExecutionId, executionStartMessage)
	if err != nil {
		log.Warningf("could not post message to case: %s", err)
	}

	err = theHiveConnector.postVariablesAsCommentInCase(
		execMetadata.ExecutionId, execMetadata.Playbook.PlaybookVariables,
		"variables at start of execution")
	if err != nil {
		log.Warningf("could not report variables in case comment: %s", err)
	}

	log.Tracef("case posted with case ID: %s", caseId)
	return string(body), nil
}

func (theHiveConnector *TheHiveConnector) UpdateEndExecutionCase(execMetadata thehive_models.ExecutionMetadata, at time.Time) (string, error) {
	caseId, err := theHiveConnector.ids_map.RetrieveCaseId(execMetadata.ExecutionId)
	if err != nil {
		return "", err
	}
	log.Trace(fmt.Sprintf("updating case status to The Hive. execution ID %s, The Hive case ID %s", execMetadata.ExecutionId, caseId))

	url := theHiveConnector.baseUrl + "/case/" + caseId
	method := "PATCH"

	err = theHiveConnector.postVariablesAsCommentInCase(execMetadata.ExecutionId, execMetadata.Variables, "variables at end of execution")
	if err != nil {
		log.Warningf("could not add task log: %s", err)
	}

	caseStatus := thehive_models.TheHiveCaseStatusTruePositive
	closureComment := fmt.Sprintf("END\nexecution ID\n\t\t[ %s ]\nended at\n\t\t[ %s ]", execMetadata.ExecutionId, at.String())
	if execMetadata.ExecutionErr != nil {
		caseStatus = thehive_models.TheHiveCaseStatusIndeterminate
		closureComment = closureComment + fmt.Sprintf("execution error: %s", execMetadata.ExecutionErr)
	}
	err = theHiveConnector.postCommentInCase(execMetadata.ExecutionId, closureComment)
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

func (theHiveConnector *TheHiveConnector) UpdateStartStepTaskInCase(execMetadata thehive_models.ExecutionMetadata, at time.Time) (string, error) {
	log.Trace(fmt.Sprintf("updating task in thehive. case ID %s. task started.", execMetadata.ExecutionId))
	taskId, err := theHiveConnector.ids_map.RetrieveTaskId(execMetadata.ExecutionId, execMetadata.Step.ID)
	if err != nil {
		return "", err
	}

	url := theHiveConnector.baseUrl + "/task/" + taskId
	method := "PATCH"

	fullyAuto := true
	for _, command := range execMetadata.Step.Commands {
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

	executionStartMessage := fmt.Sprintf(
		"START\nexecution ID\t\t[ %s ]\nstep ID\t\t[ %s ]\nstarted at\t\t[ %s ]",
		execMetadata.ExecutionId, execMetadata.Step.ID, at.String())

	err = theHiveConnector.postCommentInTaskLog(execMetadata.ExecutionId, execMetadata.Step, executionStartMessage)
	if err != nil {
		log.Warningf("could post message to task: %s", err)
	}

	err = theHiveConnector.postStepDataAsCommentInTaskLog(execMetadata.ExecutionId, execMetadata.Step, "step data")
	if err != nil {
		log.Warningf("could not report step data in step task log: %s", err)
	}

	return theHiveConnector.getIdFromRespBody(body)
}

func (theHiveConnector *TheHiveConnector) UpdateEndStepTaskInCase(execMetadata thehive_models.ExecutionMetadata, at time.Time) (string, error) {
	log.Trace(fmt.Sprintf("updating task in thehive. case ID %s. task ended.", execMetadata.ExecutionId))
	taskId, err := theHiveConnector.ids_map.RetrieveTaskId(execMetadata.ExecutionId, execMetadata.Step.ID)
	if err != nil {
		return "", err
	}

	url := theHiveConnector.baseUrl + "/task/" + taskId
	method := "PATCH"

	err = theHiveConnector.postStepVariablesAsCommentInTaskLog(execMetadata.ExecutionId, execMetadata.Step, "returned variables")
	if err != nil {
		log.Warningf("could not report variables in step task log: %s", err)
	}

	taskStatus := thehive_models.TheHiveStatusCompleted
	executionEndMessage := fmt.Sprintf(
		"END\nexecution ID\t\t[ %s ]\nstep ID\t\t[ %s ]\nended at\t\t[ %s ]",
		execMetadata.ExecutionId, execMetadata.Step.ID, at.String())

	if execMetadata.ExecutionErr != nil {
		taskStatus = thehive_models.TheHiveStatusCancelled
		executionEndMessage = executionEndMessage + fmt.Sprintf("\nexecution error: %s", execMetadata.ExecutionErr)
	}
	err = theHiveConnector.postCommentInTaskLog(execMetadata.ExecutionId, execMetadata.Step, executionEndMessage)
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

func (theHiveConnector *TheHiveConnector) Hello() string {

	url := theHiveConnector.baseUrl + "/user/current"

	body, err := theHiveConnector.sendRequest("GET", url, nil)
	if err != nil {
		return "error"
	}

	return (string(body))
}

func (theHiveConnector *TheHiveConnector) sendRequest(method string, url string, body interface{}) ([]byte, error) {
	req, err := thehive_utils.PrepareRequest(method, url, theHiveConnector.apiKey, body)
	if err != nil {
		return nil, err
	}
	return thehive_utils.SendRequest(theHiveConnector.client, req)

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
