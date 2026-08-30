package storage

import (
	"context"
	"time"

	"soarca/pkg/fins/protocol"
)

type FinStore interface {
	// Register persists a new Fin registration. Returns ErrConflict if the ID already exists.
	Create(ctx context.Context, record fin.Record) error
	// Get looks up a Fin by its FinId. Returns ErrNotFound if none exists.
	Get(ctx context.Context, finID string) (fin.Record, error)
	// GetByTokenHash looks up the Fin whose FinTokenHash matches. Returns ErrNotFound if none exists.
	GetByTokenHash(ctx context.Context, tokenHash string) (fin.Record, error)
	// List returns all registered Fins.
	List(ctx context.Context) ([]fin.Record, error)
	// Touch updates a Fin's LastSeen timestamp. Returns ErrNotFound if the Fin does not exist.
	Touch(ctx context.Context, finID string, at time.Time) error
	// Unregister removes a Fin registration. Returns ErrNotFound if none exists.
	Delete(ctx context.Context, finID string) error
}
