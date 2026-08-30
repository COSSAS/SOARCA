package storage

import (
	"context"

	"soarca/internal/playbooks"
	"soarca/pkg/cacao"
)

type PlaybookStore interface {
	// Create stores a new playbook. Returns ErrConflict if the ID already exists.
	Create(ctx context.Context, pb cacao.Playbook) error
	// Update replaces an existing playbook. Returns ErrNotFound if the ID does not exist.
	Update(ctx context.Context, pb cacao.Playbook) error
	// Get retrieves a playbook by ID. Returns ErrNotFound if it does not exist.
	Get(ctx context.Context, id string) (cacao.Playbook, error)
	// Delete removes a playbook by ID. Returns ErrNotFound if it does not exist.
	Delete(ctx context.Context, id string) error
	// List returns all stored playbooks.
	List(ctx context.Context) ([]cacao.Playbook, error)
	// ListMeta returns lightweight metadata for all stored playbooks.
	ListMeta(ctx context.Context) ([]playbooks.Meta, error)
}
