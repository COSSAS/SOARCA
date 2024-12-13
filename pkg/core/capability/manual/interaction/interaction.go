package interaction

import (
	"context"
	"fmt"
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
	GetPendingCommands() ([]manual.InteractionCommandData, error)
	// even if step has multiple manual commands, there should always be just one pending manual command per action step
	GetPendingCommand(metadata execution.Metadata) (manual.InteractionCommandData, error)
	Continue(outArgsResult manual.ManualOutArgUpdatePayload) error
}

type InteractionController struct {
	InteractionStorage map[string]map[string]manual.InteractionCommandData // Keyed on [executionID][stepID]
	Notifiers          []IInteractionIntegrationNotifier
}

func New(manualIntegrations []IInteractionIntegrationNotifier) *InteractionController {
	storage := map[string]map[string]manual.InteractionCommandData{}
	return &InteractionController{
		InteractionStorage: storage,
		Notifiers:          manualIntegrations,
	}
}

// ############################################################################
// ICapabilityInteraction implementation
// ############################################################################
func (manualController *InteractionController) Queue(command manual.InteractionCommand, manualCapabilityChannel chan manual.InteractionResponse, ctx context.Context) error {

	err := manualController.registerPendingInteraction(command)
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

			_, err := manualController.getPendingInteraction(metadata)
			if err != nil {
				// If not in there, was already resolved
				log.Warning(err)
				log.Warning("manual command not found among pending ones. should be already resolved")
				return
			}

			// Was there. It's resolved, so it's removed from the pendings register
			manualController.removeInteractionFromPending(metadata)

			interactionResponse := manual.InteractionResponse{
				ResponseError: result.ResponseError,
				Payload:       result.Payload,
			}

			manualCapabilityChannel <- interactionResponse
			return
		}
	}
}

// ############################################################################
// IInteractionStorage implementation
// ############################################################################
func (manualController *InteractionController) GetPendingCommands() ([]manual.InteractionCommandData, error) {
	log.Trace("getting pending manual commands")
	return manualController.getAllPendingInteractions(), nil
}

func (manualController *InteractionController) GetPendingCommand(metadata execution.Metadata) (manual.InteractionCommandData, error) {
	log.Trace("getting pending manual command")
	return manualController.getPendingInteraction(metadata)
}

func (manualController *InteractionController) PostContinue(outArgsResult manual.ManualOutArgUpdatePayload) error {
	log.Trace("completing manual command")
	// TODO
	// Get execution metadata from updatepayload
	// Check command is indeed pending
	// If not, it means it was already solved (right?)
	// If it is, put outArgs back into manualCapabilityChannel (must figure out how...)
	// de-register the command
	return nil
}

// ############################################################################
// Utilities and functionalities
// ############################################################################
func (manualController *InteractionController) registerPendingInteraction(command manual.InteractionCommand) error {

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
		manualController.InteractionStorage[interaction.ExecutionId] = map[string]manual.InteractionCommandData{
			interaction.StepId: interaction,
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
	execution[interaction.StepId] = interaction

	return nil
}

func (manualController *InteractionController) getAllPendingInteractions() []manual.InteractionCommandData {
	allPendingInteractions := []manual.InteractionCommandData{}
	for _, interactions := range manualController.InteractionStorage {
		for _, interaction := range interactions {
			allPendingInteractions = append(allPendingInteractions, interaction)
		}
	}
	return allPendingInteractions
}

func (manualController *InteractionController) getPendingInteraction(commandMetadata execution.Metadata) (manual.InteractionCommandData, error) {
	executionCommands, ok := manualController.InteractionStorage[commandMetadata.ExecutionId.String()]
	if !ok {
		err := fmt.Errorf("no pending commands found for execution %s", commandMetadata.ExecutionId.String())
		return manual.InteractionCommandData{}, err
	}
	commandData, ok := executionCommands[commandMetadata.StepId]
	if !ok {
		err := fmt.Errorf("no pending commands found for execution %s -> step %s",
			commandMetadata.ExecutionId.String(),
			commandMetadata.StepId,
		)
		return manual.InteractionCommandData{}, err
	}
	return commandData, nil
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
