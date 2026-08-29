package runtime

import (
	"testing"

	"soarca/internal/config"
)

func TestNewInitializesDependencies(t *testing.T) {
	runtime, err := New(Options{
		Storage: config.StorageConfig{
			UseDatabase: false,
		},
		Cache: config.CacheConfig{
			MaxExecutions: 3,
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
	if ops.Executions == nil {
		t.Fatal("Operations.Executions is nil")
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
