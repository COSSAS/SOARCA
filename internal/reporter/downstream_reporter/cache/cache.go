package cache

import (
	"errors"
	"soarca/models/cacao"
	"soarca/models/report"
	"soarca/utils"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const MaxExecutions int = 10
const MaxSteps int = 10

type Cache struct {
	Size         int
	Cache        map[string]report.ExecutionEntry // Cached up to max
	fifoRegister []string                         // Used for O(1) FIFO cache management
}

func New() *Cache {
	maxExecutions, _ := strconv.Atoi(utils.GetEnv("MAX_EXECUTIONS", strconv.Itoa(MaxExecutions)))
	return &Cache{
		Size:  maxExecutions,
		Cache: make(map[string]report.ExecutionEntry),
	}
}

func (cacheReporter *Cache) getExecution(executionKey uuid.UUID) (report.ExecutionEntry, error) {
	executionKeyStr := executionKey.String()
	executionEntry, ok := cacheReporter.Cache[executionKeyStr]
	if !ok {
		err := errors.New("execution is not in cache")
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
func (cacheReporter *Cache) AddExecution(newExecutionEntry report.ExecutionEntry) error {

	if !(len(cacheReporter.fifoRegister) == len(cacheReporter.Cache)) {
		return errors.New("cache fifo register and content are desynchronized")
	}

	newExecutionEntryKey := newExecutionEntry.ExecutionId.String()

	if len(cacheReporter.fifoRegister) >= cacheReporter.Size {
		firstExecution := cacheReporter.fifoRegister[0]
		cacheReporter.fifoRegister = cacheReporter.fifoRegister[1:]
		delete(cacheReporter.Cache, firstExecution)

		cacheReporter.Cache[newExecutionEntryKey] = newExecutionEntry
		cacheReporter.fifoRegister = append(cacheReporter.fifoRegister, newExecutionEntryKey)
		return nil
	}

	cacheReporter.fifoRegister = append(cacheReporter.fifoRegister, newExecutionEntryKey)
	cacheReporter.Cache[newExecutionEntryKey] = newExecutionEntry
	return nil
}

func (cacheReporter *Cache) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) error {
	newExecutionEntry := report.ExecutionEntry{
		ExecutionId:    executionId,
		PlaybookId:     playbook.ID,
		Started:        time.Now(),
		Ended:          time.Time{},
		StepResults:    map[string]report.StepResult{},
		PlaybookResult: nil,
		Status:         report.Ongoing,
	}
	err := cacheReporter.AddExecution(newExecutionEntry)
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

	executionEntry.Ended = time.Now()
	if workflowError != nil {
		executionEntry.PlaybookResult = workflowError
		executionEntry.Status = report.Failed
	} else {
		executionEntry.Status = report.SuccessfullyExecuted
	}

	return nil
}

func (cacheReporter *Cache) ReportStepStart(executionId uuid.UUID, step cacao.Step, variables cacao.Variables) error {
	executionEntry, err := cacheReporter.getExecution(executionId)
	if err != nil {
		return err
	}
	newStepEntry := report.StepResult{
		ExecutionId: executionId,
		StepId:      step.ID,
		Started:     time.Now(),
		Ended:       time.Time{},
		Variables:   variables,
		Status:      report.Ongoing,
		Error:       nil,
	}
	executionEntry.StepResults[step.ID] = newStepEntry
	return nil
}
func (cacheReporter *Cache) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, stepError error) error {
	executionStepResult, err := cacheReporter.getExecutionStep(executionId, step.ID)
	if err != nil {
		return err
	}

	executionStepResult.Ended = time.Now()
	if stepError != nil {
		executionStepResult.Error = stepError
		executionStepResult.Status = report.ServerSideError
	} else {
		executionStepResult.Status = report.SuccessfullyExecuted
	}
	executionStepResult.Variables = stepResults
	return nil
}

func (cacheReporter *Cache) GetExecutionReport(executionId uuid.UUID, playbook cacao.Playbook) error {
	// TODO
	return nil
}
