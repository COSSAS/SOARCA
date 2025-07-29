package mappings

import (
	"fmt"
	"reflect"
	"soarca/internal/logger"
)

var (
	component = reflect.TypeOf(ExecutionCaseMap{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// ############################### Playbook to TheHive ID mappings

type SOARCATheHiveMap struct {
	ExecutionsCaseMaps map[string]ExecutionCaseMap
}
type ExecutionCaseMap struct {
	caseId        string
	stepsTasksMap map[string]string
}

// TODO: Change to using observables instead of updating the tasks descriptions

func (soarcaTheHiveMap *SOARCATheHiveMap) CheckExecutionCaseExists(executionId string) error {
	if _, ok := soarcaTheHiveMap.ExecutionsCaseMaps[executionId]; !ok {
		return fmt.Errorf("case not found for execution id %s", executionId)
	}
	return nil
}
func (soarcaTheHiveMap *SOARCATheHiveMap) CheckExecutionStepTaskExists(executionId string, stepId string) error {
	if _, ok := soarcaTheHiveMap.ExecutionsCaseMaps[executionId].stepsTasksMap[stepId]; !ok {
		return fmt.Errorf("task not found for execution id %s for step id %s", executionId, stepId)
	}
	return nil
}

func (soarcaTheHiveMap *SOARCATheHiveMap) RegisterExecutionInCase(executionId string, caseId string) error {
	soarcaTheHiveMap.ExecutionsCaseMaps[executionId] = ExecutionCaseMap{
		caseId:        caseId,
		stepsTasksMap: map[string]string{},
	}
	log.Info(fmt.Sprintf("registering execution: %s, case id: %s", executionId, caseId))

	return nil
}
func (soarcaTheHiveMap *SOARCATheHiveMap) RegisterStepTaskInCase(executionId string, stepId string, taskId string) {
	soarcaTheHiveMap.ExecutionsCaseMaps[executionId].stepsTasksMap[stepId] = taskId
}

func (soarcaTheHiveMap *SOARCATheHiveMap) RetrieveCaseId(executionId string) (string, error) {
	err := soarcaTheHiveMap.CheckExecutionCaseExists(executionId)
	if err != nil {
		return "", err
	}
	return soarcaTheHiveMap.ExecutionsCaseMaps[executionId].caseId, nil
}

func (soarcaTheHiveMap *SOARCATheHiveMap) RetrieveTaskId(executionId string, stepId string) (string, error) {
	err := soarcaTheHiveMap.CheckExecutionCaseExists(executionId)
	if err != nil {
		return "", err
	}
	err = soarcaTheHiveMap.CheckExecutionStepTaskExists(executionId, stepId)
	if err != nil {
		return "", err
	}
	return soarcaTheHiveMap.ExecutionsCaseMaps[executionId].stepsTasksMap[stepId], nil
}
