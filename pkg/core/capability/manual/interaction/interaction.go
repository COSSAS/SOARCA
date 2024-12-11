package interaction

import (
	"fmt"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// NOTE:
// The InteractionController is injected with all configured Interactions (SOARCA API always, plus AT MOST ONE integration)
// The manual capability is injected with the InteractionController
// The manual capability triggers interactioncontroller.PostCommand
// The InteractionController register a manual command pending in its memory registry
// The manual capability waits on interactioncontroller.WasCompleted() status != pending (to implement)
// Meanwhile, external systems use the InteractionController to do GetPending. GetPending just uses the memory registry of InteractionController
// Also meanwhile, external systems can use InteractionController to do Continue()
// The manual capability continues.

type IInteractionIntegrationNotifier interface {
	Notify(command manual.InteractionIntegrationCommand, channel chan manual.InteractionIntegrationResponse)
}

type ICapabilityInteraction interface {
	Queue(command manual.InteractionCommand, channel chan manual.InteractionResponse) error
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
func (manualController *InteractionController) Queue(command manual.InteractionCommand, manualCapabilityChannel chan manual.InteractionResponse) error {

	err := manualController.registerPendingInteraction(command)
	if err != nil {
		return err
	}

	// Copy and type conversion
	integrationCommand := manual.InteractionIntegrationCommand(command)

	// One response channel for all integrations. First reply resolves the manual command
	interactionChannel := make(chan manual.InteractionIntegrationResponse)
	defer close(interactionChannel)

	for _, notifier := range manualController.Notifiers {
		go notifier.Notify(integrationCommand, interactionChannel)
	}

	// Purposedly blocking in idle-wait. We want to receive data back before continuiing the playbook
	go func() {
		for {
			// Skeleton. Implementation todo. Also study what happens if timeout at higher level
			// Also study what happens with concurrent manual commands e.g. from parallel steps,
			// with respect to using one class channel or different channels per call
			result := <-interactionChannel

			// TODO: check register for pending manual command
			// If was already resolved, safely discard
			// Otherwise, resolve command, post back to manual capability, de-register command form pending

			log.Debug(result)
		}
	}()
	return nil
}

// ############################################################################
// IInteractionStorage implementation
// ############################################################################
func (manualController *InteractionController) GetPendingCommands() ([]manual.InteractionCommandData, error) {
	log.Trace("getting pending manual commands")
	return []manual.InteractionCommandData{}, nil
}

func (manualController *InteractionController) GetPendingCommand(metadata execution.Metadata) (manual.InteractionCommandData, error) {
	log.Trace("getting pending manual command")
	return manual.InteractionCommandData{}, nil
}

func (manualController *InteractionController) PostContinue(outArgsResult manual.ManualOutArgUpdatePayload) error {
	log.Trace("completing manual command")
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

func (manualController *InteractionController) continueInteraction(interactionResponse manual.InteractionResponse) error {
	// TODO
	if interactionResponse.ResponseError != nil {
		return interactionResponse.ResponseError
	}
	return nil
}
