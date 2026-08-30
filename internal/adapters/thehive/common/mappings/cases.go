package mappings

import (
	"fmt"
	"reflect"
	"soarca/internal/logger"
)

var (
	component = reflect.TypeOf(RunCaseMap{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// ############################### Playbook to TheHive ID mappings

type SOARCATheHiveMap struct {
	RunsCaseMaps map[string]RunCaseMap
}
type RunCaseMap struct {
	caseId        string
	stepsTasksMap map[string]string
}

// TODO: Change to using observables instead of updating the tasks descriptions

func (soarcaTheHiveMap *SOARCATheHiveMap) CheckRunCaseExists(runId string) error {
	if _, ok := soarcaTheHiveMap.RunsCaseMaps[runId]; !ok {
		return fmt.Errorf("case not found for run id %s", runId)
	}
	return nil
}
func (soarcaTheHiveMap *SOARCATheHiveMap) CheckRunStepTaskExists(runId string, stepId string) error {
	if _, ok := soarcaTheHiveMap.RunsCaseMaps[runId].stepsTasksMap[stepId]; !ok {
		return fmt.Errorf("task not found for run id %s for step id %s", runId, stepId)
	}
	return nil
}

func (soarcaTheHiveMap *SOARCATheHiveMap) RegisterRunInCase(runId string, caseId string) error {
	soarcaTheHiveMap.RunsCaseMaps[runId] = RunCaseMap{
		caseId:        caseId,
		stepsTasksMap: map[string]string{},
	}
	log.Info(fmt.Sprintf("registering run: %s, case id: %s", runId, caseId))

	return nil
}
func (soarcaTheHiveMap *SOARCATheHiveMap) RegisterStepTaskInCase(runId string, stepId string, taskId string) {
	soarcaTheHiveMap.RunsCaseMaps[runId].stepsTasksMap[stepId] = taskId
}

func (soarcaTheHiveMap *SOARCATheHiveMap) RetrieveCaseId(runId string) (string, error) {
	err := soarcaTheHiveMap.CheckRunCaseExists(runId)
	if err != nil {
		return "", err
	}
	return soarcaTheHiveMap.RunsCaseMaps[runId].caseId, nil
}

func (soarcaTheHiveMap *SOARCATheHiveMap) RetrieveTaskId(runId string, stepId string) (string, error) {
	err := soarcaTheHiveMap.CheckRunCaseExists(runId)
	if err != nil {
		return "", err
	}
	err = soarcaTheHiveMap.CheckRunStepTaskExists(runId, stepId)
	if err != nil {
		return "", err
	}
	return soarcaTheHiveMap.RunsCaseMaps[runId].stepsTasksMap[stepId], nil
}
