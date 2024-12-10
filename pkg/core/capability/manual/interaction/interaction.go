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
// Upon a Continue and relative updates, the IsCompleted will return status == completed, and the related info
// The manual capability continues.

type IInteractionIntegrationNotifier interface {
	Notify(command manual.InteractionIntegrationCommandData, channel chan manual.InteractionIntegrationResponse)
}

type ICapabilityInteraction interface {
	Queue(command manual.ManualCommandData) error
}

type IInteractionStorage interface {
	GetPendingCommands() ([]manual.ManualCommandData, error)
	GetPendingCommand(metadata execution.Metadata) (manual.ManualCommandData, error)
	Continue(outArgsResult manual.ManualOutArgUpdatePayload) error
}

type InteractionController struct {
	PendingCommands map[string]manual.ManualCommandData // Keyed on execution ID
	notifiers       []IInteractionIntegrationNotifier
}

func (manualController *InteractionController) GetPendingCommands() ([]manual.ManualCommandData, error) {
	log.Trace("getting pending manual commands")
	return []manual.ManualCommandData{}, nil
}

func (manualController *InteractionController) GetPendingCommand(metadata execution.Metadata) (manual.ManualCommandData, error) {
	log.Trace("getting pending manual command")
	return manual.ManualCommandData{}, nil
}

func (manualController *InteractionController) PostContinue(outArgsResult manual.ManualOutArgUpdatePayload) error {
	log.Trace("completing manual command")
	return nil
}

func (manualController *InteractionController) Queue(command manual.InteractionIntegrationCommandData) error {
	channel := make(chan manual.InteractionIntegrationResponse)
	for _, notifier := range manualController.notifiers {
		go notifier.Notify(command, channel)
	}
	for {
		// Skeleton. Implementation todo. Also study what happens if timeout at higher level
		result := <-channel
		log.Debug(result)
		return nil
	}
}
