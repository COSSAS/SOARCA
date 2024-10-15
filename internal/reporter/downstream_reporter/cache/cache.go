package cache

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"slices"
	"soarca/models/cacao"
	cache_report "soarca/models/cache"
	itime "soarca/utils/time"
	"sync"
	"time"

	"github.com/google/uuid"
)

const MaxExecutions int = 10

type Cache struct {
	Size         int
	timeUtil     itime.ITime
	Cache        map[string]cache_report.ExecutionEntry // Cached up to max
	fifoRegister []string                               // Used for O(1) FIFO cache management
	mutex        sync.Mutex
}

func New(timeUtil itime.ITime, maxExecutions int) *Cache {
	return &Cache{
		Size:     maxExecutions,
		Cache:    make(map[string]cache_report.ExecutionEntry),
		timeUtil: timeUtil,
		mutex:    sync.Mutex{},
	}
}

// ############################### Atomic cache access operations (mutex-protection)

func (cacheReporter *Cache) getAllExecutions() ([]cache_report.ExecutionEntry, error) {
	executions := make([]cache_report.ExecutionEntry, 0)
	// NOTE: fetched via fifo register key reference as is ordered array,
	// this is needed to test and report back ordered executions stored

	// Lock
	cacheReporter.mutex.Lock()
	defer cacheReporter.mutex.Unlock()
	for _, executionEntryKey := range cacheReporter.fifoRegister {
		// NOTE: cached executions are passed by reference, so they must not be modified
		entry, ok := cacheReporter.Cache[executionEntryKey]
		if !ok {
			// Unlock
			return []cache_report.ExecutionEntry{}, errors.New("internal error. cache fifo register and cache executions mismatch")
		}
		executions = append(executions, entry)
	}

	// Unlocked
	return executions, nil
}

func (cacheReporter *Cache) getExecution(executionKey uuid.UUID) (cache_report.ExecutionEntry, error) {

	executionKeyStr := executionKey.String()
	// No need for mutex as is one-line access
	executionEntry, ok := cacheReporter.Cache[executionKeyStr]

	if !ok {
		err := errors.New("execution is not in cache. consider increasing cache size")
		return cache_report.ExecutionEntry{}, err
		// TODO Retrieve from database and push to cache
	}
	return executionEntry, nil
}

// Adding executions in FIFO logic
func (cacheReporter *Cache) addExecutionFIFO(newExecutionEntry cache_report.ExecutionEntry) error {

	if !(len(cacheReporter.fifoRegister) == len(cacheReporter.Cache)) {
		return errors.New("cache fifo register and content are desynchronized")
	}

	newExecutionEntryKey := newExecutionEntry.ExecutionId.String()

	// Lock
	cacheReporter.mutex.Lock()
	defer cacheReporter.mutex.Unlock()

	if _, ok := cacheReporter.Cache[newExecutionEntryKey]; ok {
		return errors.New("there is already an execution in the cache with the same execution id")
	}
	if len(cacheReporter.fifoRegister) >= cacheReporter.Size {

		firstExecution := cacheReporter.fifoRegister[0]
		cacheReporter.fifoRegister = cacheReporter.fifoRegister[1:]
		delete(cacheReporter.Cache, firstExecution)
		cacheReporter.fifoRegister = append(cacheReporter.fifoRegister, newExecutionEntryKey)
		cacheReporter.Cache[newExecutionEntryKey] = newExecutionEntry

		return nil
		// Unlocked
	}
	cacheReporter.fifoRegister = append(cacheReporter.fifoRegister, newExecutionEntryKey)
	cacheReporter.Cache[newExecutionEntryKey] = newExecutionEntry

	return nil
	// Unlocked
}

func (cacheReporter *Cache) upateEndExecutionWorkflow(executionId uuid.UUID, workflowError error, at time.Time) error {
	// The cache should stay locked for the whole modification period
	// in order to prevent e.g. the execution data being popped-out due to FIFO
	// while its status or some of its steps are being updated

	// Lock
	cacheReporter.mutex.Lock()
	defer cacheReporter.mutex.Unlock()

	executionEntry, err := cacheReporter.getExecution(executionId)
	if err != nil {
		return err
	}

	if workflowError != nil {
		executionEntry.Error = workflowError
		executionEntry.Status = cache_report.Failed
	} else {
		executionEntry.Status = cache_report.SuccessfullyExecuted
	}
	executionEntry.Ended = at
	cacheReporter.Cache[executionId.String()] = executionEntry

	return nil
	// Unlocked
}

func (cacheReporter *Cache) addStartExecutionStep(executionId uuid.UUID, newStepData cache_report.StepResult) error {
	// Locked
	cacheReporter.mutex.Lock()
	defer cacheReporter.mutex.Unlock()

	executionEntry, err := cacheReporter.getExecution(executionId)
	if err != nil {
		return err
	}

	if executionEntry.Status != cache_report.Ongoing {
		return errors.New("trying to report on the execution of a step for an already reportedly terminated playbook execution")
	}
	_, alreadyThere := executionEntry.StepResults[newStepData.StepId]
	if alreadyThere {
		// TODO: must fix: all steps should start empty values but already present. Check should be
		// done on Step.Started > 0 time
		//
		// Should divide between instanciation of step, and modification of step,
		// with respective checks step status
		return errors.New("a step execution start was already reported for this step. ignoring")
	}

	executionEntry.StepResults[newStepData.StepId] = newStepData
	// New code
	cacheReporter.Cache[executionId.String()] = executionEntry

	return nil
	// Unlocked
}

func (cacheReporter *Cache) upateEndExecutionStep(executionId uuid.UUID, stepId string, returnVars cacao.Variables, stepError error, acceptedStepStati []cache_report.Status, at time.Time) error {
	// Locked
	cacheReporter.mutex.Lock()
	defer cacheReporter.mutex.Unlock()

	executionEntry, err := cacheReporter.getExecution(executionId)
	if err != nil {
		return err
	}

	executionStepResult, ok := executionEntry.StepResults[stepId]
	if !ok {
		// TODO: must fix: all steps should start empty values but already present. Check should be
		// done on Step.Started > 0 time
		return errors.New("trying to update a step which was not (yet?) recorded in the cache")
		// Unlocked
	}

	if !slices.Contains(acceptedStepStati, executionStepResult.Status) {
		return fmt.Errorf("step status precondition not met for step update [step status: %s]", executionStepResult.Status.String())
	}

	if stepError != nil {
		executionStepResult.Error = stepError
		executionStepResult.Status = cache_report.ServerSideError
	} else {
		executionStepResult.Status = cache_report.SuccessfullyExecuted
	}
	executionStepResult.Ended = at
	executionStepResult.Variables = returnVars
	executionEntry.StepResults[stepId] = executionStepResult
	cacheReporter.Cache[executionId.String()] = executionEntry

	return nil
	// Unlocked
}

// ############################### Informer interface

func (cacheReporter *Cache) GetExecutions() ([]cache_report.ExecutionEntry, error) {
	executions, err := cacheReporter.getAllExecutions()
	return executions, err
}

func (cacheReporter *Cache) GetExecutionReport(executionKey uuid.UUID) (cache_report.ExecutionEntry, error) {

	executionEntry, err := cacheReporter.getExecution(executionKey)
	if err != nil {
		return cache_report.ExecutionEntry{}, err
	}
	report := executionEntry

	return report, nil
}

// ############################### Reporting interface

func (cacheReporter *Cache) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook, at time.Time) error {

	newExecutionEntry := cache_report.ExecutionEntry{
		ExecutionId: executionId,
		PlaybookId:  playbook.ID,
		Name:        playbook.Name,
		Description: playbook.Description,
		Started:     at,
		Ended:       time.Time{},
		StepResults: map[string]cache_report.StepResult{},
		Status:      cache_report.Ongoing,
	}
	err := cacheReporter.addExecutionFIFO(newExecutionEntry)
	if err != nil {
		return err
	}
	return nil
}

func (cacheReporter *Cache) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowError error, at time.Time) error {

	err := cacheReporter.upateEndExecutionWorkflow(executionId, workflowError, at)
	return err
}

func (cacheReporter *Cache) ReportStepStart(executionId uuid.UUID, step cacao.Step, variables cacao.Variables, at time.Time) error {

	commandsB64 := []string{}
	isAutomated := true
	for _, cmd := range step.Commands {
		if cmd.Type == cacao.CommandTypeManual {
			isAutomated = false
		}
		if cmd.CommandB64 != "" {
			commandsB64 = append(commandsB64, cmd.CommandB64)
		} else {
			cmdB64 := b64.StdEncoding.EncodeToString([]byte(cmd.Command))
			commandsB64 = append(commandsB64, cmdB64)
		}
	}

	newStep := cache_report.StepResult{
		ExecutionId: executionId,
		StepId:      step.ID,
		Name:        step.Name,
		Description: step.Description,
		//Started:     cacheReporter.timeUtil.Now(),
		Started:     at,
		Ended:       time.Time{},
		Variables:   variables,
		CommandsB64: commandsB64,
		Status:      cache_report.Ongoing,
		Error:       nil,
		IsAutomated: isAutomated,
	}

	err := cacheReporter.addStartExecutionStep(executionId, newStep)

	return err
}

func (cacheReporter *Cache) ReportStepEnd(executionId uuid.UUID, step cacao.Step, returnVars cacao.Variables, stepError error, at time.Time) error {

	// stepId, err := uuid.Parse(step.ID)
	// if err != nil {
	// 	return fmt.Errorf("could not parse to uuid the step id: %s", step.ID)
	// }

	acceptedStepStati := []cache_report.Status{cache_report.Ongoing}
	err := cacheReporter.upateEndExecutionStep(executionId, step.ID, returnVars, stepError, acceptedStepStati, at)

	return err
}
