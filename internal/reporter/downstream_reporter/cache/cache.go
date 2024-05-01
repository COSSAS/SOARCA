package cache

import (
	"errors"
	"fmt"
	"reflect"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/report"
	"soarca/utils"
	itime "soarca/utils/time"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var component = reflect.TypeOf(Cache{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

const MaxExecutions int = 10
const MaxSteps int = 10

type Cache struct {
	Size         int
	timeUtil     itime.ITime
	Cache        map[string]report.ExecutionEntry // Cached up to max
	fifoRegister []string                         // Used for O(1) FIFO cache management
}

func New(timeUtil itime.ITime) *Cache {
	maxExecutions, _ := strconv.Atoi(utils.GetEnv("MAX_EXECUTIONS", strconv.Itoa(MaxExecutions)))
	return &Cache{
		Size:     maxExecutions,
		Cache:    make(map[string]report.ExecutionEntry),
		timeUtil: timeUtil,
	}
}

func (cacheReporter *Cache) getExecution(executionKey uuid.UUID) (report.ExecutionEntry, error) {
	executionKeyStr := executionKey.String()
	executionEntry, ok := cacheReporter.Cache[executionKeyStr]
	if !ok {
		err := errors.New("execution is not in cache")
		log.Warning("execution is not in cache. consider increasing cache size.")
		return report.ExecutionEntry{}, err
		// TODO Retrieve from database
	}
	return executionEntry, nil
}
func (cacheReporter *Cache) getExecutionStep(executionKey uuid.UUID, stepKey string) (report.StepResult, error) {
	executionEntry, err := cacheReporter.getExecution(executionKey)
	if err != nil {
		return report.StepResult{}, err
	}
	executionStep, ok := executionEntry.StepResults[stepKey]
	if !ok {
		err := errors.New("execution step is not in cache")
		return report.StepResult{}, err
		// TODO Retrieve from database
	}
	return executionStep, nil
}

// Adding executions in FIFO logic
func (cacheReporter *Cache) addExecution(newExecutionEntry report.ExecutionEntry) error {

	if !(len(cacheReporter.fifoRegister) == len(cacheReporter.Cache)) {
		return errors.New("cache fifo register and content are desynchronized")
	}

	newExecutionEntryKey := newExecutionEntry.ExecutionId.String()

	if len(cacheReporter.fifoRegister) >= cacheReporter.Size {
		firstExecution := cacheReporter.fifoRegister[0]
		cacheReporter.fifoRegister = cacheReporter.fifoRegister[1:]
		delete(cacheReporter.Cache, firstExecution)

		cacheReporter.fifoRegister = append(cacheReporter.fifoRegister, newExecutionEntryKey)
		cacheReporter.Cache[newExecutionEntryKey] = newExecutionEntry
		return nil
	}

	cacheReporter.fifoRegister = append(cacheReporter.fifoRegister, newExecutionEntryKey)
	cacheReporter.Cache[newExecutionEntryKey] = newExecutionEntry
	return nil
}

func (cacheReporter *Cache) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) error {
	newExecutionEntry := report.ExecutionEntry{
		ExecutionId: executionId,
		PlaybookId:  playbook.ID,
		Started:     cacheReporter.timeUtil.Now(),
		Ended:       time.Time{},
		StepResults: map[string]report.StepResult{},
		Status:      report.Ongoing,
	}
	err := cacheReporter.addExecution(newExecutionEntry)
	if err != nil {
		return err
	}
	return nil
}

func (cacheReporter *Cache) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error) error {

	executionEntry, err := cacheReporter.getExecution(executionId)
	if err != nil {
		return err
	}

	if workflowError != nil {
		executionEntry.PlaybookResult = workflowError
		executionEntry.Status = report.Failed
	} else {
		executionEntry.Status = report.SuccessfullyExecuted
	}
	executionEntry.Ended = cacheReporter.timeUtil.Now()
	cacheReporter.Cache[executionId.String()] = executionEntry

	return nil
}

func (cacheReporter *Cache) ReportStepStart(executionId uuid.UUID, step cacao.Step, variables cacao.Variables) error {
	executionEntry, err := cacheReporter.getExecution(executionId)
	if err != nil {
		return err
	}

	if executionEntry.Status != report.Ongoing {
		return errors.New("trying to report on the execution of a step for an already reported completed or failed execution")
	}

	fmt.Println(executionEntry)

	_, alreadyThere := executionEntry.StepResults[step.ID]
	if alreadyThere {
		log.Warning("a step execution was already reported for this step. overwriting.")
	}

	newStepEntry := report.StepResult{
		ExecutionId: executionId,
		StepId:      step.ID,
		Started:     cacheReporter.timeUtil.Now(),
		Ended:       time.Time{},
		Variables:   variables,
		Status:      report.Ongoing,
		Error:       nil,
	}
	executionEntry.StepResults[step.ID] = newStepEntry
	return nil
}

func (cacheReporter *Cache) ReportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, stepError error) error {
	executionEntry, err := cacheReporter.getExecution(executionId)
	if err != nil {
		return err
	}

	if executionEntry.Status != report.Ongoing {
		return errors.New("trying to report on the execution of a step for an already reported completed or failed execution")
	}

	executionStepResult, err := cacheReporter.getExecutionStep(executionId, step.ID)
	if err != nil {
		return err
	}

	if executionStepResult.Status != report.Ongoing {
		return errors.New("trying to report on the execution of a step that was already reported completed or failed")
	}

	if stepError != nil {
		executionStepResult.Error = stepError
		executionStepResult.Status = report.ServerSideError
	} else {
		executionStepResult.Status = report.SuccessfullyExecuted
	}
	executionStepResult.Ended = cacheReporter.timeUtil.Now()
	executionStepResult.Variables = returnVars
	executionEntry.StepResults[step.ID] = executionStepResult

	return nil
}

func (cacheReporter *Cache) GetExecutionsIDs() []string {
	executions := make([]string, len(cacheReporter.fifoRegister))
	_ = copy(executions, cacheReporter.fifoRegister)
	return executions
}

func (cacheReporter *Cache) GetExecutionReport(executionKey uuid.UUID) (report.ExecutionEntry, error) {
	executionEntry, err := cacheReporter.getExecution(executionKey)
	if err != nil {
		return report.ExecutionEntry{}, err
	}
	report := executionEntry

	return report, nil
}
