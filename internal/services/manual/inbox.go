package manual

import (
	"soarca/pkg/core/capability/manual/interaction"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"
)

// Inbox implements the services.ManualInbox interface.
// It delegates manual interaction storage operations to the interaction controller.
type Inbox struct {
	storage interaction.IInteractionStorage
}

// NewInbox creates a manual inbox service.
func NewInbox(storage interaction.IInteractionStorage) *Inbox {
	return &Inbox{storage: storage}
}

// ListPendingCommands returns all pending manual commands.
func (i *Inbox) ListPendingCommands() ([]manual.CommandInfo, error) {
	return i.storage.GetPendingCommands()
}

// GetPendingCommand returns one pending manual command.
func (i *Inbox) GetPendingCommand(metadata execution.Metadata) (manual.CommandInfo, error) {
	return i.storage.GetPendingCommand(metadata)
}

// ContinuePendingCommand submits a response for a pending manual command.
func (i *Inbox) ContinuePendingCommand(response manual.InteractionResponse) error {
	return i.storage.PostContinue(response)
}
