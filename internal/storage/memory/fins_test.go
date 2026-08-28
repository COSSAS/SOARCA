package memory

import (
	"context"
	"testing"
	"time"

	"soarca/internal/storage"
	"soarca/pkg/models/fin"
)

func newTestFin(finID string) fin.Record {
	return fin.Record{
		FinId:         finID,
		FinTokenHash:  "hash-" + finID,
		LastSeen:      time.Now(),
	}
}

func TestFinCreate(t *testing.T) {
	fs := newFinStore()
	record := newTestFin("fin-1")
	ctx := context.Background()

	err := fs.Create(ctx, record)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Should fail if already exists
	err = fs.Create(ctx, record)
	if err != storage.ErrConflict {
		t.Errorf("Create() on duplicate should return ErrConflict, got %v", err)
	}
}

func TestFinGet(t *testing.T) {
	fs := newFinStore()
	record := newTestFin("fin-1")
	ctx := context.Background()

	// Should not exist yet
	_, err := fs.Get(ctx, "fin-1")
	if err != storage.ErrNotFound {
		t.Errorf("Get() on missing should return ErrNotFound, got %v", err)
	}

	// Create and retrieve
	fs.Create(ctx, record)
	retrieved, err := fs.Get(ctx, "fin-1")
	if err != nil {
		t.Fatalf("Get() after Create() failed: %v", err)
	}
	if retrieved.FinId != record.FinId {
		t.Errorf("Retrieved fin ID mismatch: %s != %s", retrieved.FinId, record.FinId)
	}
}

func TestFinGetByTokenHash(t *testing.T) {
	fs := newFinStore()
	record := newTestFin("fin-1")
	ctx := context.Background()

	// Should not exist yet
	_, err := fs.GetByTokenHash(ctx, "hash-fin-1")
	if err != storage.ErrNotFound {
		t.Errorf("GetByTokenHash() on missing should return ErrNotFound, got %v", err)
	}

	// Create and find by hash
	fs.Create(ctx, record)
	retrieved, err := fs.GetByTokenHash(ctx, "hash-fin-1")
	if err != nil {
		t.Fatalf("GetByTokenHash() failed: %v", err)
	}
	if retrieved.FinId != record.FinId {
		t.Errorf("Retrieved fin ID mismatch: %s != %s", retrieved.FinId, record.FinId)
	}
}

func TestFinList(t *testing.T) {
	fs := newFinStore()
	ctx := context.Background()

	fs.Create(ctx, newTestFin("fin-1"))
	fs.Create(ctx, newTestFin("fin-2"))

	list, err := fs.List(ctx)
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List() returned %d items, expected 2", len(list))
	}
}

func TestFinTouch(t *testing.T) {
	fs := newFinStore()
	record := newTestFin("fin-1")
	ctx := context.Background()

	// Should fail if doesn't exist
	now := time.Now()
	err := fs.Touch(ctx, "fin-1", now)
	if err != storage.ErrNotFound {
		t.Errorf("Touch() on missing should return ErrNotFound, got %v", err)
	}

	// Create then touch
	fs.Create(ctx, record)
	newTime := time.Now().Add(1 * time.Hour)
	err = fs.Touch(ctx, "fin-1", newTime)
	if err != nil {
		t.Fatalf("Touch() failed: %v", err)
	}

	retrieved, _ := fs.Get(ctx, "fin-1")
	if retrieved.LastSeen != newTime {
		t.Errorf("Touch did not update LastSeen")
	}
}

func TestFinDelete(t *testing.T) {
	fs := newFinStore()
	record := newTestFin("fin-1")
	ctx := context.Background()

	// Should fail if doesn't exist
	err := fs.Delete(ctx, "fin-1")
	if err != storage.ErrNotFound {
		t.Errorf("Delete() on missing should return ErrNotFound, got %v", err)
	}

	// Create then unregister
	fs.Create(ctx, record)
	err = fs.Delete(ctx, "fin-1")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Should be gone
	_, err = fs.Get(ctx, "fin-1")
	if err != storage.ErrNotFound {
		t.Errorf("Get() after Delete() should return ErrNotFound, got %v", err)
	}
}
