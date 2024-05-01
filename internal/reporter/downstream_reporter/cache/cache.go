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

func (cacheReporter *Cache) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, err error) error {

	executionEntry, ok := cacheReporter.Cache[executionId.String()]
	if !ok {
		return errors.New("execution is not in cache.")
		// TODO Retrieve from database
	}

	executionEntry.Ended = time.Now()

	if err != nil {
		executionEntry.PlaybookResult = err
		executionEntry.Status = report.Failed
	} else {
		executionEntry.Status = report.SuccessfullyExecuted
	}

	return nil
}

func (cacheReporter *Cache) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error) error {
	// TODO
	return nil
}
func (cacheReporter *Cache) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error) error {
	// TODO
	return nil
}

func (cacheReporter *Cache) GetExecutionReport(executionId uuid.UUID, playbook cacao.Playbook) error {
	// TODO
	return nil
}
