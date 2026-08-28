package memory

import (
	"context"
	"testing"

	"soarca/internal/storage"
)

func TestStoreInterfaces(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}

	if s.Playbooks() == nil {
		t.Error("Playbooks() returned nil")
	}

	if s.Fins() == nil {
		t.Error("Fins() returned nil")
	}

	err := s.Close(context.Background())
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestPlaybooksInterface(t *testing.T) {
	var _ storage.PlaybookStore = (*PlaybookStore)(nil)
}

func TestFinsInterface(t *testing.T) {
	var _ storage.FinStore = (*FinStore)(nil)
}
