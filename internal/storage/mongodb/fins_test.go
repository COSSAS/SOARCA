package mongodb

import (
	"context"
	"testing"
	"time"

	"soarca/internal/storage"
	"soarca/pkg/models/fin"
)

func newTestFin(finID string) fin.Record {
	return fin.Record{
		FinId:        finID,
		FinTokenHash: "hash-" + finID,
		LastSeen:     time.Now(),
	}
}

func TestMongoFinCreate(t *testing.T) {
	t.Skip("Requires running MongoDB instance; run with -run=TestMongoFin* after starting mongo")
	uri := skipIfNoMongo(t)
	cfg := Config{
		URI: uri,
	}

	store, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close(context.Background())

	fs := store.Fins()
	record := newTestFin("fin-1")
	ctx := context.Background()

	err = fs.Create(ctx, record)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Should fail if already exists
	err = fs.Create(ctx, record)
	if err != storage.ErrConflict {
		t.Errorf("Create() on duplicate should return ErrConflict, got %v", err)
	}
}

func TestMongoFinGet(t *testing.T) {
	t.Skip("Requires running MongoDB instance")
	uri := skipIfNoMongo(t)
	cfg := Config{
		URI: uri,
	}

	store, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close(context.Background())

	fs := store.Fins()
	record := newTestFin("fin-1")
	ctx := context.Background()

	// Should not exist yet
	_, err = fs.Get(ctx, "fin-1")
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

func TestMongoFinGetByTokenHash(t *testing.T) {
	t.Skip("Requires running MongoDB instance")
	uri := skipIfNoMongo(t)
	cfg := Config{
		URI: uri,
	}

	store, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close(context.Background())

	fs := store.Fins()
	record := newTestFin("fin-1")
	ctx := context.Background()

	// Should not exist yet
	_, err = fs.GetByTokenHash(ctx, "hash-fin-1")
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

func TestMongoFinTouch(t *testing.T) {
	t.Skip("Requires running MongoDB instance")
	uri := skipIfNoMongo(t)
	cfg := Config{
		URI: uri,
	}

	store, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close(context.Background())

	fs := store.Fins()
	record := newTestFin("fin-1")
	ctx := context.Background()

	// Should fail if doesn't exist
	now := time.Now()
	err = fs.Touch(ctx, "fin-1", now)
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
	if retrieved.LastSeen.Unix() != newTime.Unix() {
		t.Errorf("Touch did not update LastSeen")
	}
}

func TestMongoFinDelete(t *testing.T) {
	t.Skip("Requires running MongoDB instance")
	uri := skipIfNoMongo(t)
	cfg := Config{
		URI: uri,
	}

	store, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close(context.Background())

	fs := store.Fins()
	record := newTestFin("fin-1")
	ctx := context.Background()

	// Should fail if doesn't exist
	err = fs.Delete(ctx, "fin-1")
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
