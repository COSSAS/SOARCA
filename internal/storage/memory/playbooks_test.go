package memory

import (
	"context"
	"testing"

	"soarca/internal/storage"
	"soarca/pkg/models/cacao"
)

func newTestPlaybook(id string) cacao.Playbook {
	return cacao.Playbook{
		ID:   id,
		Name: "Test Playbook",
	}
}

func TestPlaybookCreate(t *testing.T) {
	ps := newPlaybookStore()
	pb := newTestPlaybook("test-1")
	ctx := context.Background()

	err := ps.Create(ctx, pb)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Should fail if already exists
	err = ps.Create(ctx, pb)
	if err != storage.ErrConflict {
		t.Errorf("Create() on duplicate should return ErrConflict, got %v", err)
	}
}

func TestPlaybookGet(t *testing.T) {
	ps := newPlaybookStore()
	pb := newTestPlaybook("test-1")
	ctx := context.Background()

	// Should not exist yet
	_, err := ps.Get(ctx, "test-1")
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

func TestPlaybookUpdate(t *testing.T) {
	ps := newPlaybookStore()
	pb := newTestPlaybook("test-1")
	ctx := context.Background()

	// Should fail if doesn't exist
	err := ps.Update(ctx, pb)
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

func TestPlaybookDelete(t *testing.T) {
	ps := newPlaybookStore()
	pb := newTestPlaybook("test-1")
	ctx := context.Background()

	// Should fail if doesn't exist
	err := ps.Delete(ctx, "test-1")
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

func TestPlaybookList(t *testing.T) {
	ps := newPlaybookStore()
	ctx := context.Background()

	ps.Create(ctx, newTestPlaybook("test-1"))
	ps.Create(ctx, newTestPlaybook("test-2"))

	list, err := ps.List(ctx)
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List() returned %d items, expected 2", len(list))
	}
}

func TestPlaybookListMeta(t *testing.T) {
	ps := newPlaybookStore()
	pb := newTestPlaybook("test-1")
	pb.Description = "Test Description"
	ctx := context.Background()

	ps.Create(ctx, pb)

	list, err := ps.ListMeta(ctx)
	if err != nil {
		t.Fatalf("ListMeta() failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListMeta() returned %d items, expected 1", len(list))
	}
	if list[0].ID != "test-1" {
		t.Errorf("Meta ID mismatch: %s", list[0].ID)
	}
	if list[0].Description != "Test Description" {
		t.Errorf("Meta Description mismatch: %s", list[0].Description)
	}
}
