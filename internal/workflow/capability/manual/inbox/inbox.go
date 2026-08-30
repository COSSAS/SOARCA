package inbox

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/logger"
	"soarca/internal/registry"
	"soarca/pkg/cacao"
	"soarca/internal/manual/model"
	"soarca/internal/runs/model"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type Notifier interface {
	Notify(command manual.Notification, channel chan manual.Response)
}

type Dispatcher interface {
	Queue(command manual.CommandInfo, manualComms manual.Waiter) error
	// Deregister removes the pending interaction for metadata, if still
	// present. Callers that own the full lifecycle of a queued command
	// (i.e. they know when it has been resolved or has timed out) should
	// call this synchronously as soon as that happens, so a subsequent
	// re-run of the same step - e.g. a step inside a while-loop body -
	// never races an async cleanup routine trying to remove the same,
	// already-superseded entry.
	Deregister(metadata run.Metadata) error
}

type Store interface {
	GetPendingCommands() ([]manual.CommandInfo, error)
	// GetPendingCommand looks up one specific pending command by its
	// StepRunId. Multiple pending commands may legitimately share the
	// same StepId - e.g. overlapping while-loop iterations, or (once
	// implemented) concurrent parallel branches converging on the same
	// step - so StepId alone cannot identify a single pending command; it's
	// up to the caller (UI/integrator) to disambiguate between several
	// pending commands for the same StepId using StepRunId.
	GetPendingCommand(metadata run.Metadata) (manual.CommandInfo, error)
	PostContinue(response manual.Response) error
}

type Inbox struct {
	pending   *registry.Registry[manual.PendingCommand] // Keyed on [runID][stepRunID]
	Notifiers []Notifier
}

func New(manualIntegrations []Notifier) *Inbox {
	return &Inbox{
		pending:   registry.New[manual.PendingCommand](),
		Notifiers: manualIntegrations,
	}
}

// ############################################################################
// Dispatcher implementation
// ############################################################################
func (manualController *Inbox) Queue(command manual.CommandInfo, manualComms manual.Waiter) error {

	err := manualController.registerPendingCommand(command, manualComms.Channel)
	if err != nil {
		return err
	}

	if _, ok := manualComms.TimeoutContext.Deadline(); !ok {
		return errors.New("manual command does not have a deadline")
	}

	// Copy and type conversion
	integrationCommand := manual.Notification(command)

	// One response channel for all integrations
	integrationChannel := make(chan manual.Response)

	for _, notifier := range manualController.Notifiers {
		go notifier.Notify(integrationCommand, integrationChannel)
	}

	// Backstop cleanup: removes the pending interaction if the caller
	// driving this command (e.g. ManualCapability.Execute) never calls
	// Deregister itself - for instance if Queue is used directly, as in
	// this package's own tests. If Deregister has already been called
	// synchronously by the caller, this is a harmless no-op (see
	// handleManualCommandResponse).
	go manualController.handleManualCommandResponse(command, manualComms)

	return nil
}

// Deregister removes the pending interaction for metadata, if still
// present. It is safe to call even if the interaction was already
// removed (e.g. by the internal timeout/completion cleanup goroutine
// racing this call): in that case it returns
// manual.ErrorPendingCommandNotFound, which callers can treat as a
// benign, expected outcome rather than a failure.
func (manualController *Inbox) Deregister(metadata run.Metadata) error {
	return manualController.removeCommandFromPending(metadata)
}

func (manualController *Inbox) handleManualCommandResponse(command manual.CommandInfo, manualComms manual.Waiter) {
	log.Trace(
		fmt.Sprintf(
			"goroutine handling command response %s, %s (step run %s) has started",
			command.Metadata.RunId.String(), command.Metadata.StepId, command.Metadata.StepRunId.String()))
	defer log.Trace(
		fmt.Sprintf(
			"goroutine handling command response %s, %s (step run %s) has ended",
			command.Metadata.RunId.String(), command.Metadata.StepId, command.Metadata.StepRunId.String()))

	// Wait for either timeout or response
	<-manualComms.TimeoutContext.Done()
	if manualComms.TimeoutContext.Err() == context.DeadlineExceeded {
		log.Info("manual command timed out. deregistering associated pending command")
	} else if manualComms.TimeoutContext.Err() == context.Canceled {
		log.Info("manual command completed. deregistering associated pending command")
	}
	err := manualController.removeCommandFromPending(command.Metadata)
	if err != nil {
		log.Warning(err)
		log.Warning("manual command not found among pending ones. should be already resolved")
		return
	}
}

// ############################################################################
// Store implementation
// ############################################################################
func (manualController *Inbox) GetPendingCommands() ([]manual.CommandInfo, error) {
	log.Trace("getting pending manual commands")
	return manualController.getAllPendingCommandsInfo(), nil
}

func (manualController *Inbox) GetPendingCommand(metadata run.Metadata) (manual.CommandInfo, error) {
	log.Trace("getting pending manual command")
	pending, err := manualController.getPendingCommand(metadata)
	return pending.CommandInfo, err
}

func (manualController *Inbox) PostContinue(response manual.Response) error {
	log.Trace("completing manual command")

	// If not in there, it means it was already solved, or expired
	pendingEntry, err := manualController.getPendingCommand(response.Metadata)
	if err != nil {
		log.Warning(err)
		return err
	}

	warnings, err := manualController.validateMatchingOutArgs(pendingEntry, response.OutArgsVariables)
	if err != nil {
		return err
	}

	//Then put outArgs back into manualCapabilityChannel
	// Copy result and conversion back to response format
	log.Trace("pushing assigned variables in manual capability channel")
	pendingEntry.Channel <- response

	if len(warnings) > 0 {
		for _, warning := range warnings {
			log.Warning(warning)
		}
	}

	return nil
}

// ############################################################################
// Utilities and functionalities
// ############################################################################
func (manualController *Inbox) registerPendingCommand(command manual.CommandInfo, manualChan chan manual.Response) error {

	commandInfo := manual.CommandInfo{
		Metadata:         command.Metadata,
		Context:          command.Context,
		OutArgsVariables: command.OutArgsVariables,
	}

	entry := manual.PendingCommand{
		CommandInfo: commandInfo,
		Channel:     manualChan,
	}

	err := manualController.pending.Register(
		commandInfo.Metadata.RunId.String(),
		commandInfo.Metadata.StepRunId.String(),
		entry,
	)
	if err != nil {
		var alreadyRegistered registry.ErrAlreadyRegistered
		if errors.As(err, &alreadyRegistered) {
			// Practically unreachable in normal operation: StepRunId
			// is a fresh UUID minted once per step invocation
			// (decomposer.newStepMetadata), so a collision here means Queue
			// was called twice for the exact same invocation.
			err := fmt.Errorf(
				"a manual command is already pending for run %s, step run %s (step %s)",
				commandInfo.Metadata.RunId.String(), commandInfo.Metadata.StepRunId.String(), commandInfo.Metadata.StepId)
			log.Error(err)
			return err
		}
		return err
	}

	return nil
}

func (manualController *Inbox) getAllPendingCommandsInfo() []manual.CommandInfo {
	entries := manualController.pending.List()
	allPendingCommands := make([]manual.CommandInfo, 0, len(entries))
	for _, entry := range entries {
		allPendingCommands = append(allPendingCommands, entry.CommandInfo)
	}
	return allPendingCommands
}

func (manualController *Inbox) getPendingCommand(commandMetadata run.Metadata) (manual.PendingCommand, error) {
	entry, err := manualController.pending.Get(commandMetadata.RunId.String(), commandMetadata.StepRunId.String())
	if err != nil {
		var outerNotFound registry.ErrOuterKeyNotFound
		if errors.As(err, &outerNotFound) {
			errMsg := fmt.Sprintf("no pending commands found for run %s", commandMetadata.RunId.String())
			return manual.PendingCommand{}, manual.ErrorPendingCommandNotFound{Err: errMsg}
		}
		var innerNotFound registry.ErrInnerKeyNotFound
		if errors.As(err, &innerNotFound) {
			errMsg := fmt.Sprintf("no pending command found for run %s -> step run %s",
				commandMetadata.RunId.String(),
				commandMetadata.StepRunId.String(),
			)
			return manual.PendingCommand{}, manual.ErrorPendingCommandNotFound{Err: errMsg}
		}
		return manual.PendingCommand{}, err
	}
	return entry, nil
}

func (manualController *Inbox) removeCommandFromPending(commandMetadata run.Metadata) error {
	_, err := manualController.getPendingCommand(commandMetadata)
	if err != nil {
		return err
	}
	// Errors from Remove are already covered by the getPendingCommand
	// check above, so this pair (run id, step run id) is known
	// to exist.
	return manualController.pending.Remove(commandMetadata.RunId.String(), commandMetadata.StepRunId.String())
}

func (manualController *Inbox) validateMatchingOutArgs(pendingEntry manual.PendingCommand, responseOutArgs cacao.Variables) ([]string, error) {
	warns := []string{}
	var err error = nil
	for varName, variable := range responseOutArgs {
		// first check that out args provided match the variables
		if _, ok := pendingEntry.CommandInfo.OutArgsVariables[varName]; !ok {
			err = fmt.Errorf("provided out arg %s does not match any intended out arg", varName)
			return warns, manual.ErrorNonMatchingOutArgs{Err: err.Error()}

		}
		// then warn if any value outside "value" has changed
		if pending, ok := pendingEntry.CommandInfo.OutArgsVariables[varName]; ok {
			if variable.Constant != pending.Constant {
				warns = append(warns, fmt.Sprintf("provided out arg %s has different value for 'Constant' property of intended out arg. This different value is ignored.", varName))
			}
			if variable.Description != pending.Description {
				warns = append(warns, fmt.Sprintf("provided out arg %s has different value for 'Description' property of intended out arg. This different value is ignored.", varName))
			}
			if variable.External != pending.External {
				warns = append(warns, fmt.Sprintf("provided out arg %s has different value for 'External' property of intended out arg. This different value is ignored.", varName))
			}
			if variable.Type != pending.Type {
				warns = append(warns, fmt.Sprintf("provided out arg %s has different value for 'Type' property of intended out arg. This different value is ignored.", varName))
			}
		}
	}
	return warns, err
}
