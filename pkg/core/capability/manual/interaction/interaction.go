package interaction

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/core/registry"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type IInteractionIntegrationNotifier interface {
	Notify(command manual.InteractionIntegrationCommand, channel chan manual.InteractionResponse)
}

type ICapabilityInteraction interface {
	Queue(command manual.CommandInfo, manualComms manual.ManualCapabilityCommunication) error
	// Deregister removes the pending interaction for metadata, if still
	// present. Callers that own the full lifecycle of a queued command
	// (i.e. they know when it has been resolved or has timed out) should
	// call this synchronously as soon as that happens, so a subsequent
	// re-execution of the same step - e.g. a step inside a while-loop body -
	// never races an async cleanup routine trying to remove the same,
	// already-superseded entry.
	Deregister(metadata execution.Metadata) error
}

type IInteractionStorage interface {
	GetPendingCommands() ([]manual.CommandInfo, error)
	// GetPendingCommand looks up one specific pending command by its
	// StepExecutionId. Multiple pending commands may legitimately share the
	// same StepId - e.g. overlapping while-loop iterations, or (once
	// implemented) concurrent parallel branches converging on the same
	// step - so StepId alone cannot identify a single pending command; it's
	// up to the caller (UI/integrator) to disambiguate between several
	// pending commands for the same StepId using StepExecutionId.
	GetPendingCommand(metadata execution.Metadata) (manual.CommandInfo, error)
	PostContinue(response manual.InteractionResponse) error
}

type InteractionController struct {
	pending   *registry.Registry[manual.InteractionStorageEntry] // Keyed on [executionID][stepExecutionID]
	Notifiers []IInteractionIntegrationNotifier
}

func New(manualIntegrations []IInteractionIntegrationNotifier) *InteractionController {
	return &InteractionController{
		pending:   registry.New[manual.InteractionStorageEntry](),
		Notifiers: manualIntegrations,
	}
}

// ############################################################################
// ICapabilityInteraction implementation
// ############################################################################
func (manualController *InteractionController) Queue(command manual.CommandInfo, manualComms manual.ManualCapabilityCommunication) error {

	err := manualController.registerPendingInteraction(command, manualComms.Channel)
	if err != nil {
		return err
	}

	if _, ok := manualComms.TimeoutContext.Deadline(); !ok {
		return errors.New("manual command does not have a deadline")
	}

	// Copy and type conversion
	integrationCommand := manual.InteractionIntegrationCommand(command)

	// One response channel for all integrations
	integrationChannel := make(chan manual.InteractionResponse)

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
func (manualController *InteractionController) Deregister(metadata execution.Metadata) error {
	return manualController.removeInteractionFromPending(metadata)
}

func (manualController *InteractionController) handleManualCommandResponse(command manual.CommandInfo, manualComms manual.ManualCapabilityCommunication) {
	log.Trace(
		fmt.Sprintf(
			"goroutine handling command response %s, %s (step execution %s) has started",
			command.Metadata.ExecutionId.String(), command.Metadata.StepId, command.Metadata.StepExecutionId.String()))
	defer log.Trace(
		fmt.Sprintf(
			"goroutine handling command response %s, %s (step execution %s) has ended",
			command.Metadata.ExecutionId.String(), command.Metadata.StepId, command.Metadata.StepExecutionId.String()))

	// Wait for either timeout or response
	<-manualComms.TimeoutContext.Done()
	if manualComms.TimeoutContext.Err() == context.DeadlineExceeded {
		log.Info("manual command timed out. deregistering associated pending command")
	} else if manualComms.TimeoutContext.Err() == context.Canceled {
		log.Info("manual command completed. deregistering associated pending command")
	}
	err := manualController.removeInteractionFromPending(command.Metadata)
	if err != nil {
		log.Warning(err)
		log.Warning("manual command not found among pending ones. should be already resolved")
		return
	}
}

// ############################################################################
// IInteractionStorage implementation
// ############################################################################
func (manualController *InteractionController) GetPendingCommands() ([]manual.CommandInfo, error) {
	log.Trace("getting pending manual commands")
	return manualController.getAllPendingCommandsInfo(), nil
}

func (manualController *InteractionController) GetPendingCommand(metadata execution.Metadata) (manual.CommandInfo, error) {
	log.Trace("getting pending manual command")
	interaction, err := manualController.getPendingInteraction(metadata)
	return interaction.CommandInfo, err
}

func (manualController *InteractionController) PostContinue(response manual.InteractionResponse) error {
	log.Trace("completing manual command")

	// If not in there, it means it was already solved, or expired
	pendingEntry, err := manualController.getPendingInteraction(response.Metadata)
	if err != nil {
		log.Warning(err)
		return err
	}

	warnings, err := manualController.validateMatchingOutArgs(pendingEntry, response.OutArgsVariables)
	if err != nil {
		return err
	}

	//Then put outArgs back into manualCapabilityChannel
	// Copy result and conversion back to interactionResponse format
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
func (manualController *InteractionController) registerPendingInteraction(command manual.CommandInfo, manualChan chan manual.InteractionResponse) error {

	commandInfo := manual.CommandInfo{
		Metadata:         command.Metadata,
		Context:          command.Context,
		OutArgsVariables: command.OutArgsVariables,
	}

	entry := manual.InteractionStorageEntry{
		CommandInfo: commandInfo,
		Channel:     manualChan,
	}

	err := manualController.pending.Register(
		commandInfo.Metadata.ExecutionId.String(),
		commandInfo.Metadata.StepExecutionId.String(),
		entry,
	)
	if err != nil {
		var alreadyRegistered registry.ErrAlreadyRegistered
		if errors.As(err, &alreadyRegistered) {
			// Practically unreachable in normal operation: StepExecutionId
			// is a fresh UUID minted once per step invocation
			// (decomposer.newStepMetadata), so a collision here means Queue
			// was called twice for the exact same invocation.
			err := fmt.Errorf(
				"a manual command is already pending for execution %s, step execution %s (step %s)",
				commandInfo.Metadata.ExecutionId.String(), commandInfo.Metadata.StepExecutionId.String(), commandInfo.Metadata.StepId)
			log.Error(err)
			return err
		}
		return err
	}

	return nil
}

func (manualController *InteractionController) getAllPendingCommandsInfo() []manual.CommandInfo {
	entries := manualController.pending.List()
	allPendingInteractions := make([]manual.CommandInfo, 0, len(entries))
	for _, entry := range entries {
		allPendingInteractions = append(allPendingInteractions, entry.CommandInfo)
	}
	return allPendingInteractions
}

func (manualController *InteractionController) getPendingInteraction(commandMetadata execution.Metadata) (manual.InteractionStorageEntry, error) {
	entry, err := manualController.pending.Get(commandMetadata.ExecutionId.String(), commandMetadata.StepExecutionId.String())
	if err != nil {
		var outerNotFound registry.ErrOuterKeyNotFound
		if errors.As(err, &outerNotFound) {
			errMsg := fmt.Sprintf("no pending commands found for execution %s", commandMetadata.ExecutionId.String())
			return manual.InteractionStorageEntry{}, manual.ErrorPendingCommandNotFound{Err: errMsg}
		}
		var innerNotFound registry.ErrInnerKeyNotFound
		if errors.As(err, &innerNotFound) {
			errMsg := fmt.Sprintf("no pending command found for execution %s -> step execution %s",
				commandMetadata.ExecutionId.String(),
				commandMetadata.StepExecutionId.String(),
			)
			return manual.InteractionStorageEntry{}, manual.ErrorPendingCommandNotFound{Err: errMsg}
		}
		return manual.InteractionStorageEntry{}, err
	}
	return entry, nil
}

func (manualController *InteractionController) removeInteractionFromPending(commandMetadata execution.Metadata) error {
	_, err := manualController.getPendingInteraction(commandMetadata)
	if err != nil {
		return err
	}
	// Errors from Remove are already covered by the getPendingInteraction
	// check above, so this pair (execution id, step execution id) is known
	// to exist.
	return manualController.pending.Remove(commandMetadata.ExecutionId.String(), commandMetadata.StepExecutionId.String())
}

func (manualController *InteractionController) validateMatchingOutArgs(pendingEntry manual.InteractionStorageEntry, responseOutArgs cacao.Variables) ([]string, error) {
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
