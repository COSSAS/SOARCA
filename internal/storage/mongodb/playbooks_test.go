package mongodb

import (
	"context"
	"testing"

	"soarca/internal/storage"
	"soarca/pkg/models/cacao"
)

func skipIfNoMongo(t *testing.T) string {
	return "mongodb://localhost:27017/soarca_test"
}

func newTestPlaybook(id string) cacao.Playbook {
	return cacao.Playbook{
		ID:   id,
		Name: "Test Playbook",
	}
}

func TestMongoPlaybookCreate(t *testing.T) {
	t.Skip("Requires running MongoDB instance; run with -run=TestMongoPlaybook* after starting mongo")
	uri := skipIfNoMongo(t)
	cfg := Config{
		URI:            uri,
		
	}

	store, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close(context.Background())

	ps := store.Playbooks()
	pb := newTestPlaybook("test-1")
	ctx := context.Background()

	err = ps.Create(ctx, pb)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Should fail if already exists
	err = ps.Create(ctx, pb)
	if err != storage.ErrConflict {
		t.Errorf("Create() on duplicate should return ErrConflict, got %v", err)
	}
}

func TestMongoPlaybookGet(t *testing.T) {
	t.Skip("Requires running MongoDB instance")
	uri := skipIfNoMongo(t)
	cfg := Config{
		URI:            uri,
		
	}

	store, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close(context.Background())

	ps := store.Playbooks()
	pb := newTestPlaybook("test-1")
	ctx := context.Background()

	// Should not exist yet
	_, err = ps.Get(ctx, "test-1")
	if err != storage.ErrNotFound {
		t.Errorf("Get() on missing should return ErrNotFound, got %v", err)
	}

	// Create and retrieve
	ps.Create(ctx, pb)
	retrieved, err := ps.Get(ctx, "test-1")
	if err != nil {
		t.Fatalf("Get() after Create() failed: %v", err)
	}
	if retrieved.ID != pb.ID {
		t.Errorf("Retrieved playbook ID mismatch: %s != %s", retrieved.ID, pb.ID)
	}
}

func TestMongoPlaybookUpdate(t *testing.T) {
	t.Skip("Requires running MongoDB instance")
	uri := skipIfNoMongo(t)
	cfg := Config{
		URI:            uri,
		
	}

	store, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close(context.Background())

	ps := store.Playbooks()
	pb := newTestPlaybook("test-1")
	ctx := context.Background()

	// Should fail if doesn't exist
	err = ps.Update(ctx, pb)
	if err != storage.ErrNotFound {
		t.Errorf("Update() on missing should return ErrNotFound, got %v", err)
	}

	// Create then update
	ps.Create(ctx, pb)
	pb.Name = "Updated Name"
	err = ps.Update(ctx, pb)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	retrieved, _ := ps.Get(ctx, "test-1")
	if retrieved.Name != "Updated Name" {
		t.Errorf("Update did not persist: %s", retrieved.Name)
	}
}

func TestMongoPlaybookDelete(t *testing.T) {
	t.Skip("Requires running MongoDB instance")
	uri := skipIfNoMongo(t)
	cfg := Config{
		URI:            uri,
		
	}

	store, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer store.Close(context.Background())

	ps := store.Playbooks()
	pb := newTestPlaybook("test-1")
	ctx := context.Background()

	// Should fail if doesn't exist
	err = ps.Delete(ctx, "test-1")
	if err != storage.ErrNotFound {
		t.Errorf("Delete() on missing should return ErrNotFound, got %v", err)
	}

	// Create then delete
	ps.Create(ctx, pb)
	err = ps.Delete(ctx, "test-1")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Should be gone
	_, err = ps.Get(ctx, "test-1")
	if err != storage.ErrNotFound {
		t.Errorf("Get() after Delete() should return ErrNotFound, got %v", err)
	}
}
