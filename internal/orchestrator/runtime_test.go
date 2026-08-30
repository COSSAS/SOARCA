package orchestrator

import (
	"testing"

	"soarca/internal/config"
)

func TestNewInitializesDependencies(t *testing.T) {
	runtime, err := New(Options{
		Storage: config.StorageConfig{
			DatabaseURL: "sqlite://:memory:",
		},
		RunState: config.RunStateConfig{
			MaxRuns: 3,
		},
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := runtime.Close(); err != nil {
			t.Fatalf("Close() returned error: %v", err)
		}
	})

	ops := runtime.Operations()

	if ops.Playbooks == nil {
		t.Fatal("Operations.Playbooks is nil")
	}
	if ops.Runs == nil {
		t.Fatal("Operations.Runs is nil")
	}
	if ops.Fins == nil {
		t.Fatal("Operations.Fins is nil")
	}
	if ops.Work == nil {
		t.Fatal("Operations.Work is nil")
	}
	if ops.Manual == nil {
		t.Fatal("Operations.Manual is nil")
	}
}
