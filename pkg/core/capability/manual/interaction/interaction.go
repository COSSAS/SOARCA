package interaction

import (
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
	InteractionStorage map[string]manual.InteractionCommandData // Keyed on execution ID
	Notifiers          []IInteractionIntegrationNotifier
}

func New(manualIntegrations []IInteractionIntegrationNotifier) *InteractionController {
	storage := map[string]manual.InteractionCommandData{}
	return &InteractionController{
		InteractionStorage: storage,
		Notifiers:          manualIntegrations,
	}
}

// ############################################################################
// ICapabilityInteraction implementation
// ############################################################################
func (manualController *InteractionController) Queue(command manual.InteractionCommand, manualCapabilityChannel chan manual.InteractionResponse) error {

	// Note: there is one manual capability per whole execution, which means there is one manualCapabilityChannel per execution
	//

	// TODO regsiter pending command in storage

	integrationCommand := manual.InteractionIntegrationCommand{
		Metadata: command.Metadata,
		Context:  command.Context,
	}

	// One response channel for all integrations. First reply resolves the manual command
	interactionChannel := make(chan manual.InteractionIntegrationResponse)

	for _, notifier := range manualController.Notifiers {
		go notifier.Notify(integrationCommand, interactionChannel)
	}

	// Purposedly blocking in idle-wait. We want to receive data back before continuiing the playbook
	for {
		// Skeleton. Implementation todo. Also study what happens if timeout at higher level
		// Also study what happens with concurrent manual commands e.g. from parallel steps,
		// with respect to using one class channel or different channels per call
		result := <-interactionChannel

		// TODO: check register for pending manual command
		// If was already resolved, safely discard
		// Otherwise, resolve command, post back to manual capability, de-register command form pending

		log.Debug(result)
		return nil
	}
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
func (manualController *InteractionController) registerPendingInteraction() {
	// TODO
}

func (manualController *InteractionController) continueInteraction() {
	// TODO
}
