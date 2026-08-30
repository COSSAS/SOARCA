package runstate

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"slices"
	run "soarca/internal/runs/model"
	runstate_report "soarca/internal/runs/state"
	"soarca/pkg/cacao"
	itime "soarca/pkg/utils/time"
	"sync"
	"time"

	"github.com/google/uuid"
)

const MaxRuns int = 10

type RunState struct {
	Size         int
	timeUtil     itime.ITime
	runs         map[string]runstate_report.RunEntry
	fifoRegister []string
	mutex        sync.Mutex
}

func New(timeUtil itime.ITime, maxRuns int) *RunState {
	return &RunState{
		Size:     maxRuns,
		runs:     make(map[string]runstate_report.RunEntry),
		timeUtil: timeUtil,
		mutex:    sync.Mutex{},
	}
}

// ############################### Atomic runstate access operations (mutex-protection)

func (runStateReporter *RunState) getAllRuns() ([]runstate_report.RunEntry, error) {
	runs := make([]runstate_report.RunEntry, 0)
	// NOTE: fetched via fifo register key reference as is ordered array,
	// this is needed to test and report back ordered runs stored

	// Lock
	runStateReporter.mutex.Lock()
	defer runStateReporter.mutex.Unlock()
	for _, runEntryKey := range runStateReporter.fifoRegister {
		// NOTE: stored runs are passed by reference, so they must not be modified
		entry, ok := runStateReporter.runs[runEntryKey]
		if !ok {
			// Unlock
			return []runstate_report.RunEntry{}, errors.New("internal error. runstate fifo register and runstate runs mismatch")
		}
		runs = append(runs, entry)
	}

	// Unlocked
	return runs, nil
}

func (runStateReporter *RunState) getRun(runKey uuid.UUID) (runstate_report.RunEntry, error) {

	runKeyStr := runKey.String()
	// No need for mutex as is one-line access
	runEntry, ok := runStateReporter.runs[runKeyStr]

	if !ok {
		err := errors.New("run is not in runstate. consider increasing runstate size")
		return runstate_report.RunEntry{}, err
		// TODO Retrieve from database and push to runstate
	}
	return runEntry, nil
}

// Adding runs in FIFO logic
func (runStateReporter *RunState) addRunFIFO(newRunEntry runstate_report.RunEntry) error {

	if len(runStateReporter.fifoRegister) != len(runStateReporter.runs) {
		return errors.New("runstate fifo register and content are desynchronized")
	}

	newRunEntryKey := newRunEntry.RunId.String()

	// Lock
	runStateReporter.mutex.Lock()
	defer runStateReporter.mutex.Unlock()

	if _, ok := runStateReporter.runs[newRunEntryKey]; ok {
		return errors.New("there is already an run in the runstate with the same run id")
	}
	if len(runStateReporter.fifoRegister) >= runStateReporter.Size {

		firstRun := runStateReporter.fifoRegister[0]
		runStateReporter.fifoRegister = runStateReporter.fifoRegister[1:]
		delete(runStateReporter.runs, firstRun)
		runStateReporter.fifoRegister = append(runStateReporter.fifoRegister, newRunEntryKey)
		runStateReporter.runs[newRunEntryKey] = newRunEntry

		return nil
		// Unlocked
	}
	runStateReporter.fifoRegister = append(runStateReporter.fifoRegister, newRunEntryKey)
	runStateReporter.runs[newRunEntryKey] = newRunEntry

	return nil
	// Unlocked
}

func (runStateReporter *RunState) upateEndRunWorkflow(runId uuid.UUID, workflowError error, at time.Time) error {
	// The runstate should stay locked for the whole modification period
	// in order to prevent e.g. the run data being popped-out due to FIFO
	// while its status or some of its steps are being updated

	// Lock
	runStateReporter.mutex.Lock()
	defer runStateReporter.mutex.Unlock()

	runEntry, err := runStateReporter.getRun(runId)
	if err != nil {
		return err
	}

	if workflowError != nil {
		runEntry.Error = workflowError
		runEntry.Status = runstate_report.Failed
	} else {
		runEntry.Status = runstate_report.SuccessfullyExecuted
	}
	runEntry.Ended = at
	runStateReporter.runs[runId.String()] = runEntry

	return nil
	// Unlocked
}

func (runStateReporter *RunState) addStartRunStep(runId uuid.UUID, newStepData runstate_report.StepResult) error {
	// Locked
	runStateReporter.mutex.Lock()
	defer runStateReporter.mutex.Unlock()

	runEntry, err := runStateReporter.getRun(runId)
	if err != nil {
		return err
	}

	if runEntry.Status != runstate_report.Ongoing {
		return errors.New("trying to report on the run of a step for an already reportedly terminated playbook run")
	}
	stepRunKey := newStepData.StepRunId.String()
	_, alreadyThere := runEntry.StepResults[stepRunKey]
	if alreadyThere {
		// A collision here would mean the same StepRunId was minted
		// twice, which should never happen - each step invocation gets a
		// fresh one. Re-runs of the same StepId are expected and get
		// their own distinct entry.
		return errors.New("a step run start was already reported for this step run. ignoring")
	}

	runEntry.StepResults[stepRunKey] = newStepData
	// New code
	runStateReporter.runs[runId.String()] = runEntry

	return nil
	// Unlocked
}

func (runStateReporter *RunState) upateEndRunStep(runId uuid.UUID, stepRunId uuid.UUID, returnVars cacao.Variables, stepError error, acceptedStepStati []runstate_report.Status, at time.Time) error {
	// Locked
	runStateReporter.mutex.Lock()
	defer runStateReporter.mutex.Unlock()

	runEntry, err := runStateReporter.getRun(runId)
	if err != nil {
		return err
	}

	stepRunKey := stepRunId.String()
	runStepResult, ok := runEntry.StepResults[stepRunKey]
	if !ok {
		return errors.New("trying to update a step run which was not (yet?) recorded in the runstate")
		// Unlocked
	}

	if !slices.Contains(acceptedStepStati, runStepResult.Status) {
		return fmt.Errorf("step status precondition not met for step update [step status: %s]", runStepResult.Status.String())
	}

	if stepError != nil {
		runStepResult.Error = stepError
		runStepResult.Status = runstate_report.ServerSideError
	} else {
		runStepResult.Status = runstate_report.SuccessfullyExecuted
	}
	runStepResult.Ended = at
	runStepResult.Variables = returnVars
	runEntry.StepResults[stepRunKey] = runStepResult
	runStateReporter.runs[runId.String()] = runEntry

	return nil
	// Unlocked
}

// Run-state query interface

func (runStateReporter *RunState) GetRuns() ([]runstate_report.RunEntry, error) {
	runs, err := runStateReporter.getAllRuns()
	return runs, err
}

func (runStateReporter *RunState) GetRunReport(runKey uuid.UUID) (runstate_report.RunEntry, error) {

	runEntry, err := runStateReporter.getRun(runKey)
	if err != nil {
		return runstate_report.RunEntry{}, err
	}

	return runEntry, nil
}

// ############################### Reporting interface

func (runStateReporter *RunState) ReportWorkflowStart(runId uuid.UUID, playbook cacao.Playbook, at time.Time) error {

	newRunEntry := runstate_report.RunEntry{
		RunId:       runId,
		PlaybookId:  playbook.ID,
		Name:        playbook.Name,
		Description: playbook.Description,
		Started:     at,
		Ended:       time.Time{},
		StepResults: map[string]runstate_report.StepResult{},
		Status:      runstate_report.Ongoing,
	}
	err := runStateReporter.addRunFIFO(newRunEntry)
	if err != nil {
		return err
	}
	return nil
}

func (runStateReporter *RunState) ReportWorkflowEnd(runId uuid.UUID, playbook cacao.Playbook, workflowError error, at time.Time) error {

	err := runStateReporter.upateEndRunWorkflow(runId, workflowError, at)
	return err
}

func (runStateReporter *RunState) ReportStepStart(metadata run.Metadata, step cacao.Step, variables cacao.Variables, at time.Time) error {

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

	newStep := runstate_report.StepResult{
		RunId:       metadata.RunId,
		StepId:      step.ID,
		StepRunId:   metadata.StepRunId,
		Name:        step.Name,
		Description: step.Description,
		//Started:     runStateReporter.timeUtil.Now(),
		Started:     at,
		Ended:       time.Time{},
		Variables:   variables,
		CommandsB64: commandsB64,
		Status:      runstate_report.Ongoing,
		Error:       nil,
		IsAutomated: isAutomated,
	}

	err := runStateReporter.addStartRunStep(metadata.RunId, newStep)

	return err
}

func (runStateReporter *RunState) ReportStepEnd(metadata run.Metadata, step cacao.Step, returnVars cacao.Variables, stepError error, at time.Time) error {

	acceptedStepStati := []runstate_report.Status{runstate_report.Ongoing}
	err := runStateReporter.upateEndRunStep(metadata.RunId, metadata.StepRunId, returnVars, stepError, acceptedStepStati, at)

	return err
}
