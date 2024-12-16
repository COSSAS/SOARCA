package interaction

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"

	"github.com/google/uuid"
)

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
	Queue(command manual.InteractionCommand, channel chan manual.InteractionResponse, ctx context.Context) error
}

type IInteractionStorage interface {
	GetPendingCommands() ([]manual.InteractionCommandData, int, error)
	// even if step has multiple manual commands, there should always be just one pending manual command per action step
	GetPendingCommand(metadata execution.Metadata) (manual.InteractionCommandData, int, error)
	Continue(outArgsResult manual.ManualOutArgUpdatePayload) (int, error)
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
func (manualController *InteractionController) Queue(command manual.InteractionCommand, manualCapabilityChannel chan manual.InteractionResponse, ctx context.Context) error {

	err := manualController.registerPendingInteraction(command, manualCapabilityChannel)
	if err != nil {
		return err
	}

	// Copy and type conversion
	integrationCommand := manual.InteractionIntegrationCommand(command)

	// One response channel for all integrations
	interactionChannel := make(chan manual.InteractionIntegrationResponse)
	defer close(interactionChannel)

	for _, notifier := range manualController.Notifiers {
		go notifier.Notify(integrationCommand, interactionChannel)
	}

	// Async idle wait on interaction integration channel
	go manualController.waitInteractionIntegrationResponse(manualCapabilityChannel, ctx, interactionChannel)

	return nil
}

func (manualController *InteractionController) waitInteractionIntegrationResponse(manualCapabilityChannel chan manual.InteractionResponse, ctx context.Context, interactionChannel chan manual.InteractionIntegrationResponse) {
	defer close(interactionChannel)
	for {
		select {
		case <-ctx.Done():
			log.Debug("context canceled due to timeout. exiting goroutine")
			return

		case result := <-interactionChannel:
			// Check register for pending manual command
			metadata := execution.Metadata{
				ExecutionId: uuid.MustParse(result.Payload.ExecutionId),
				PlaybookId:  result.Payload.PlaybookId,
				StepId:      result.Payload.StepId,
			}

			// Remove interaction from pending ones
			err := manualController.removeInteractionFromPending(metadata)
			if err != nil {
				// If it was not there, was already resolved
				log.Warning(err)
				log.Warning("manual command not found among pending ones. should be already resolved")
				return
			}

			// Copy result and conversion back to interactionResponse format
			interactionResponse := manual.InteractionResponse(result)

			manualCapabilityChannel <- interactionResponse
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

func (manualController *InteractionController) PostContinue(outArgsResult manual.ManualOutArgUpdatePayload) (int, error) {
	log.Trace("completing manual command")

	metadata := execution.Metadata{
		ExecutionId: uuid.MustParse(outArgsResult.ExecutionId),
		PlaybookId:  outArgsResult.PlaybookId,
		StepId:      outArgsResult.StepId,
	}

	// TODO: determine status code

	// If not, it means it was already solved (right?)
	pendingEntry, err := manualController.getPendingInteraction(metadata)
	if err != nil {
		log.Warning(err)
		return http.StatusAlreadyReported, err
	}

	// If it is, put outArgs back into manualCapabilityChannel

	pendingEntry.Channel <- manual.InteractionResponse{
		ResponseError: nil,
		Payload:       outArgsResult,
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
		OutArgs:       command.Context.Variables,
	}

	execution, ok := manualController.InteractionStorage[interaction.ExecutionId]

	if !ok {
		// It's fine, no entry for execution registered. Register execution and step entry
		manualController.InteractionStorage[interaction.ExecutionId] = map[string]manual.InteractionStorageEntry{
			interaction.StepId: manual.InteractionStorageEntry{
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

// func (manualController *InteractionController) continueInteraction(interactionResponse manual.InteractionResponse) error {
// 	// TODO
// 	if interactionResponse.ResponseError != nil {
// 		return interactionResponse.ResponseError
// 	}
// 	return nil
// }
