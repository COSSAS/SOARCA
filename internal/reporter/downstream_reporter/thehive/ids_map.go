package thehive

import (
	"fmt"
)

// ############################### Playbook to TheHive ID mappings

type SOARCATheHiveMap struct {
	executionsCaseMaps map[string]ExecutionCaseMap
}
type ExecutionCaseMap struct {
	caseId        string
	stepsTasksMap map[string]string
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

func (soarcaTheHiveMap *SOARCATheHiveMap) registerExecutionInCase(executionId string, caseId string) error {
	soarcaTheHiveMap.executionsCaseMaps[executionId] = ExecutionCaseMap{
		caseId:        caseId,
		stepsTasksMap: map[string]string{},
	}
	//fmt.Printf("registering execution: %s, case id: %s", executionId, caseId)
	//fmt.Printf("execution entry id %s", soarcaTheHiveMap.executionsCaseMaps[executionId].caseId)
	log.Debugf("registering execution: %s, case id: %s", executionId, caseId)
	log.Debugf("execution entry id %s", soarcaTheHiveMap.executionsCaseMaps[executionId].caseId)

	return nil
}
func (soarcaTheHiveMap *SOARCATheHiveMap) registerStepTaskInCase(executionId string, stepId string, taskId string) {
	soarcaTheHiveMap.executionsCaseMaps[executionId].stepsTasksMap[stepId] = taskId
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

// func (soarcaTheHiveMap *SOARCATheHiveMap) clearCase(executionId string) error {
// 	err := soarcaTheHiveMap.checkExecutionCaseExists(executionId)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (soarcaTheHiveMap *SOARCATheHiveMap) clearMap(executionId string) error
