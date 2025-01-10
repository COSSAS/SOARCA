package interaction

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"

	"github.com/google/uuid"
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
	Notify(command manual.InteractionIntegrationCommand, channel chan manual.InteractionIntegrationResponse)
}

type ICapabilityInteraction interface {
	Queue(command manual.InteractionCommand, manualComms manual.ManualCapabilityCommunication) error
}

type IInteractionStorage interface {
	GetPendingCommands() ([]manual.InteractionCommandData, int, error)
	// even if step has multiple manual commands, there should always be just one pending manual command per action step
	GetPendingCommand(metadata execution.Metadata) (manual.InteractionCommandData, int, error)
	PostContinue(outArgsResult manual.ManualOutArgsUpdatePayload) (int, error)
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

// ############################################################################
// ICapabilityInteraction implementation
// ############################################################################
func (manualController *InteractionController) Queue(command manual.InteractionCommand, manualComms manual.ManualCapabilityCommunication) error {

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
	integrationChannel := make(chan manual.InteractionIntegrationResponse)

	for _, notifier := range manualController.Notifiers {
		go notifier.Notify(integrationCommand, integrationChannel)
	}

	// Async idle wait on interaction integration channel
	go manualController.waitInteractionIntegrationResponse(manualComms, integrationChannel)

	return nil
}

func (manualController *InteractionController) waitInteractionIntegrationResponse(manualComms manual.ManualCapabilityCommunication, integrationChannel chan manual.InteractionIntegrationResponse) {
	defer close(integrationChannel)
	for {
		select {
		case <-manualComms.TimeoutContext.Done():
			log.Info("context canceled due to response or timeout. exiting goroutine")
			return

		case <-manualComms.Channel:
			log.Info("detected activity on manual capability channel. exiting goroutine without consuming the message")
			return

		case result := <-integrationChannel:
			// Check register for pending manual command
			metadata, err := manualController.makeExecutionMetadataFromPayload(result.Payload)
			if err != nil {
				log.Error(err)
				manualComms.Channel <- manual.InteractionResponse{
					ResponseError: err,
					Payload:       cacao.Variables{},
				}
				return
			}
			// Remove interaction from pending ones
			err = manualController.removeInteractionFromPending(metadata)
			if err != nil {
				// If it was not there, was already resolved
				log.Warning(err)
				// Captured if channel not yet closed
				log.Warning("manual command not found among pending ones. should be already resolved")
				manualComms.Channel <- manual.InteractionResponse{
					ResponseError: err,
					Payload:       cacao.Variables{},
				}
				return
			}

			// Copy result and conversion back to interactionResponse format
			returnedVars := manualController.copyOutArgsToVars(result.Payload.ResponseOutArgs)

			interactionResponse := manual.InteractionResponse{
				ResponseError: result.ResponseError,
				Payload:       returnedVars,
			}

			manualComms.Channel <- interactionResponse
			return
		}
	}
}

// ############################################################################
// IInteractionStorage implementation
// ############################################################################
func (manualController *InteractionController) GetPendingCommands() ([]manual.InteractionCommandData, int, error) {
	log.Trace("getting pending manual commands")
	return manualController.getAllPendingInteractions(), http.StatusOK, nil
}

func (manualController *InteractionController) GetPendingCommand(metadata execution.Metadata) (manual.InteractionCommandData, int, error) {
	log.Trace("getting pending manual command")
	interaction, err := manualController.getPendingInteraction(metadata)
	// TODO: determine status code
	return interaction.CommandData, http.StatusOK, err
}

func (manualController *InteractionController) PostContinue(result manual.ManualOutArgsUpdatePayload) (int, error) {
	log.Trace("completing manual command")

	metadata, err := manualController.makeExecutionMetadataFromPayload(result)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// If not in there, it means it was already solved (right?)
	pendingEntry, err := manualController.getPendingInteraction(metadata)
	if err != nil {
		log.Warning(err)
		return http.StatusAlreadyReported, err
	}

	// If it is
	for varName, variable := range result.ResponseOutArgs {
		// first check that out args provided match the variables
		if _, ok := pendingEntry.CommandData.OutVariables[varName]; !ok {
			err := errors.New("provided out args do not match command-related variables")
			log.Warning("provided out args do not match command-related variables")
			return http.StatusBadRequest, err
		}
		// then warn if any value outside "value" has changed
		if pending, ok := pendingEntry.CommandData.OutVariables[varName]; ok {
			if variable.Constant != pending.Constant {
				log.Warningf("provided out arg %s is attempting to change 'Constant' property", varName)
			}
			if variable.Description != pending.Description {
				log.Warningf("provided out arg %s is attempting to change 'Description' property", varName)
			}
			if variable.External != pending.External {
				log.Warningf("provided out arg %s is attempting to change 'External' property", varName)
			}
			if variable.Type != pending.Type {
				log.Warningf("provided out arg %s is attempting to change 'Type' property", varName)
			}
		}
	}

	//Then put outArgs back into manualCapabilityChannel
	// Copy result and conversion back to interactionResponse format
	returnedVars := manualController.copyOutArgsToVars(result.ResponseOutArgs)
	log.Trace("pushing assigned variables in manual capability channel")
	pendingEntry.Channel <- manual.InteractionResponse{
		ResponseError: nil,
		Payload:       returnedVars,
	}
	// de-register the command
	err = manualController.removeInteractionFromPending(metadata)
	if err != nil {
		log.Error(err)
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

// ############################################################################
// Utilities and functionalities
// ############################################################################
func (manualController *InteractionController) registerPendingInteraction(command manual.InteractionCommand, manualChan chan manual.InteractionResponse) error {

	interaction := manual.InteractionCommandData{
		Type:          command.Context.Command.Type,
		ExecutionId:   command.Metadata.ExecutionId.String(),
		PlaybookId:    command.Metadata.PlaybookId,
		StepId:        command.Metadata.StepId,
		Description:   command.Context.Command.Description,
		Command:       command.Context.Command.Command,
		CommandBase64: command.Context.Command.CommandB64,
		Target:        command.Context.Target,
		OutVariables:  command.Context.Variables.Select(command.Context.Step.OutArgs),
	}

	execution, ok := manualController.InteractionStorage[interaction.ExecutionId]

	if !ok {
		// It's fine, no entry for execution registered. Register execution and step entry
		manualController.InteractionStorage[interaction.ExecutionId] = map[string]manual.InteractionStorageEntry{
			interaction.StepId: {
				CommandData: interaction,
				Channel:     manualChan,
			},
		}
		return nil
	}

	// There is an execution entry
	if _, ok := execution[interaction.StepId]; ok {
		// Error: there is already a pending manual command for the action step
		err := fmt.Errorf(
			"a manual step is already pending for execution %s, step %s. There can only be one pending manual command per action step.",
			interaction.ExecutionId, interaction.StepId)
		log.Error(err)
		return err
	}

	// Execution exist, and Finally register pending command in existing execution
	// Question: is it ever the case that the same exact step is executed in parallel branches? Then this code would not work
	execution[interaction.StepId] = manual.InteractionStorageEntry{
		CommandData: interaction,
		Channel:     manualChan,
	}

	return nil
}

func (manualController *InteractionController) getAllPendingInteractions() []manual.InteractionCommandData {
	allPendingInteractions := []manual.InteractionCommandData{}
	for _, interactions := range manualController.InteractionStorage {
		for _, interaction := range interactions {
			allPendingInteractions = append(allPendingInteractions, interaction.CommandData)
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

func (manualController *InteractionController) copyOutArgsToVars(outArgs manual.ManualOutArgs) cacao.Variables {
	vars := cacao.NewVariables()
	for name, outVar := range outArgs {

		vars[name] = cacao.Variable{
			Type:  outVar.Type,
			Name:  outVar.Name,
			Value: outVar.Value,
		}
	}
	return vars
}

func (manualController *InteractionController) makeExecutionMetadataFromPayload(payload manual.ManualOutArgsUpdatePayload) (execution.Metadata, error) {
	executionId, err := uuid.Parse(payload.ExecutionId)
	if err != nil {
		return execution.Metadata{}, err
	}
	metadata := execution.Metadata{
		ExecutionId: executionId,
		PlaybookId:  payload.PlaybookId,
		StepId:      payload.StepId,
	}
	return metadata, nil
}
