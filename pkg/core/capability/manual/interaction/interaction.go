package interaction

import (
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"
	ctxModel "soarca/pkg/models/utils/context"
)

// TODO
// Add manual capability to action execution,

// NOTE: current outArgs management for Manual commands:
//	- The decomposer passes the PlaybookStepMetadata object to the
//		action executor, which includes Step
// 	- The action executor calls Execute on the capability (command type)
//		passing capability.Context, which includes the Step object
//	- The manual capability calls Queue passing InteractionCommand,
//		which includes capability.Context
// 	- Queue() posts a message, which shall include the text of the manual command,
//		and the varibales (outArgs) expected
// 	- registerPendingInteraction records the CACAO Variables corresponding to the
//		outArgs field (in the step. In future, in the command)
//	- A manual response posts back a map[string]manual.ManualOutArg object,
//		which is exactly like cacao variables, but with different requested fields.
// 	- The Interaction object cleans the returned variables to only keep
//		the name, type, and value (to not overwrite other fields)

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
}

type IInteractionStorage interface {
	GetPendingCommands() ([]manual.CommandInfo, error)
	// even if step has multiple manual commands, there should always be just one pending manual command per action step
	GetPendingCommand(metadata execution.Metadata) (manual.CommandInfo, error)
	PostContinue(response manual.InteractionResponse) error
}

type InteractionController struct {
	InteractionStorage map[string]map[string]manual.InteractionStorageEntry // Keyed on [executionID][stepID]
	Notifiers          []IInteractionIntegrationNotifier
}

func New(manualIntegrations []IInteractionIntegrationNotifier) *InteractionController {
	storage := map[string]map[string]manual.InteractionStorageEntry{}
	return &InteractionController{
		InteractionStorage: storage,
		Notifiers:          manualIntegrations,
	}
}

// TODO:
// - Add check on timeoutcontext.Done() for timeout (vs completion), and remove entry from pending in that case
// - Change waitInteractionIntegrationResponse to be waitResponse
// - Put result := <- interactionintegrationchannel into a separate function
// - Just use the one instance of manual capability channel. Do not use interactionintegrationchannel
// - Create typed error and pass back to API function for Storage interface fcns

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

	// Async idle wait on command-specific channel closure
	go manualController.handleManualCommandResponse(command, manualComms)

	return nil
}

func (manualController *InteractionController) handleManualCommandResponse(command manual.CommandInfo, manualComms manual.ManualCapabilityCommunication) {
	log.Trace(
		fmt.Sprintf(
			"goroutine handling command response %s, %s has started", command.Metadata.ExecutionId.String(), command.Metadata.StepId))
	defer log.Trace(
		fmt.Sprintf(
			"goroutine handling command response %s, %s has ended", command.Metadata.ExecutionId.String(), command.Metadata.StepId))

	select {
	case <-manualComms.TimeoutContext.Done():
		if manualComms.TimeoutContext.Err().Error() == ctxModel.ErrorContextTimeout {
			log.Info("manual command timed out. deregistering associated pending command")

			err := manualController.removeInteractionFromPending(command.Metadata)
			if err != nil {
				log.Warning(err)
				log.Warning("manual command not found among pending ones. should be already resolved")
				return
			}
		} else if manualComms.TimeoutContext.Err().Error() == ctxModel.ErrorContextCanceled {
			log.Info("manual command completed. deregistering associated pending command")
			err := manualController.removeInteractionFromPending(command.Metadata)
			if err != nil {
				log.Warning(err)
				log.Warning("manual command not found among pending ones. should be already resolved")
				return
			}
		}
	}
}

// ############################################################################
// IInteractionStorage implementation
// ############################################################################
func (manualController *InteractionController) GetPendingCommands() ([]manual.CommandInfo, error) {
	log.Trace("getting pending manual commands")
	return manualController.getAllPendingInteractions(), nil
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

	execution, ok := manualController.InteractionStorage[commandInfo.Metadata.ExecutionId.String()]

	if !ok {
		// It's fine, no entry for execution registered. Register execution and step entry
		manualController.InteractionStorage[commandInfo.Metadata.ExecutionId.String()] = map[string]manual.InteractionStorageEntry{
			commandInfo.Metadata.StepId: {
				CommandInfo: commandInfo,
				Channel:     manualChan,
			},
		}
		return nil
	}

	// There is an execution entry
	if _, ok := execution[commandInfo.Metadata.StepId]; ok {
		// Error: there is already a pending manual command for the action step
		err := fmt.Errorf(
			"a manual step is already pending for execution %s, step %s. There can only be one pending manual command per action step.",
			commandInfo.Metadata.ExecutionId.String(), commandInfo.Metadata.StepId)
		log.Error(err)
		return err
	}

	// Execution exist, and Finally register pending command in existing execution
	// Question: is it ever the case that the same exact step is executed in parallel branches? Then this code would not work
	execution[commandInfo.Metadata.StepId] = manual.InteractionStorageEntry{
		CommandInfo: commandInfo,
		Channel:     manualChan,
	}

	return nil
}

func (manualController *InteractionController) getAllPendingInteractions() []manual.CommandInfo {
	allPendingInteractions := []manual.CommandInfo{}
	for _, interactions := range manualController.InteractionStorage {
		for _, interaction := range interactions {
			allPendingInteractions = append(allPendingInteractions, interaction.CommandInfo)
		}
	}
	return allPendingInteractions
}

func (manualController *InteractionController) getPendingInteraction(commandMetadata execution.Metadata) (manual.InteractionStorageEntry, error) {
	executionCommands, ok := manualController.InteractionStorage[commandMetadata.ExecutionId.String()]
	if !ok {
		err := fmt.Errorf("no pending commands found for execution %s", commandMetadata.ExecutionId.String())
		return manual.InteractionStorageEntry{}, err
	}
	interaction, ok := executionCommands[commandMetadata.StepId]
	if !ok {
		err := fmt.Errorf("no pending commands found for execution %s -> step %s",
			commandMetadata.ExecutionId.String(),
			commandMetadata.StepId,
		)
		return manual.InteractionStorageEntry{}, err
	}
	return interaction, nil
}

func (manualController *InteractionController) removeInteractionFromPending(commandMetadata execution.Metadata) error {
	_, err := manualController.getPendingInteraction(commandMetadata)
	if err != nil {
		return err
	}
	// Get map of pending manual commands associated to execution
	executionCommands := manualController.InteractionStorage[commandMetadata.ExecutionId.String()]
	// Delete stepID-linked pending command
	delete(executionCommands, commandMetadata.StepId)

	// If no pending commands associated to the execution, delete the executions map
	// This is done to keep the storage clean.
	if len(executionCommands) == 0 {
		delete(manualController.InteractionStorage, commandMetadata.ExecutionId.String())
	}
	return nil
}

func (manualController *InteractionController) validateMatchingOutArgs(pendingEntry manual.InteractionStorageEntry, responseOutArgs cacao.Variables) ([]string, error) {
	warns := []string{}
	var err error = nil
	for varName, variable := range responseOutArgs {
		// first check that out args provided match the variables
		if _, ok := pendingEntry.CommandInfo.OutArgsVariables[varName]; !ok {
			err = errors.New(fmt.Sprintf("provided out arg %s does not match any intended out arg", varName))
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
