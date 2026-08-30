package manual

import (
	"soarca/internal/workflow/capability/manual/inbox"
	"soarca/internal/manual/model"
	"soarca/internal/runs/model"
)

// Inbox implements the services.ManualInbox interface.
// It delegates manual manual inbox storage operations to the manual inbox.
type Inbox struct {
	storage inbox.Store
}

// NewInbox creates a manual inbox service.
func NewInbox(storage inbox.Store) *Inbox {
	return &Inbox{storage: storage}
}

// ListPendingCommands returns all pending manual commands.
func (i *Inbox) ListPendingCommands() ([]manual.CommandInfo, error) {
	return i.storage.GetPendingCommands()
}

// GetPendingCommand returns one pending manual command.
func (i *Inbox) GetPendingCommand(metadata run.Metadata) (manual.CommandInfo, error) {
	return i.storage.GetPendingCommand(metadata)
}

// ContinuePendingCommand submits a response for a pending manual command.
func (i *Inbox) ContinuePendingCommand(response manual.Response) error {
	return i.storage.PostContinue(response)
}
