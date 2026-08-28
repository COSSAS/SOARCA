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

	if runtime.PlaybookStore == nil {
		t.Fatal("PlaybookStore is nil")
	}
	if runtime.FinStore == nil {
		t.Fatal("FinStore is nil")
	}
	if runtime.Cache == nil {
		t.Fatal("Cache is nil")
	}
	if runtime.Interaction == nil {
		t.Fatal("Interaction is nil")
	}
	if runtime.FinQueue == nil {
		t.Fatal("FinQueue is nil")
	}
}
