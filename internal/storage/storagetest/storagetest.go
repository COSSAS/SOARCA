// Package storagetest provides a disposable store for tests.
package storagetest

import (
	"context"
	"testing"

	"soarca/internal/storage"
	storagesql "soarca/internal/storage/sql"
)

// New returns a store backed by a SQLite database private to this test, so
// tests exercise the same SQL path as production.
func New(t *testing.T) storage.Store {
	t.Helper()

	store, err := storagesql.New(context.Background(),
		"file:"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("storagetest.New() returned error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close(context.Background()) })
	return store
}
