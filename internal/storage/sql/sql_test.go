package sql

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"soarca/internal/storage"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/fin"
)

// newTestStore gives each test its own private in-memory database.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := New(context.Background(), "file:"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close(context.Background()) })
	return store
}

func testPlaybook(id string) cacao.Playbook {
	return cacao.Playbook{
		ID:          id,
		Name:        "Test Playbook",
		Description: "a playbook",
		ValidFrom:   time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		ValidUntil:  time.Date(2124, 1, 1, 9, 0, 0, 0, time.UTC),
		Labels:      []string{"soarca", "test"},
		Workflow: cacao.Workflow{
			"start--test": cacao.Step{ID: "start--test", Type: cacao.StepTypeStart},
		},
	}
}

func TestPlaybookCreateAndGet(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	pb := testPlaybook("playbook--1")

	if err := store.Playbooks().Create(ctx, pb); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	got, err := store.Playbooks().Get(ctx, pb.ID)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if got.ID != pb.ID || got.Name != pb.Name {
		t.Errorf("Get() = %s/%s, want %s/%s", got.ID, got.Name, pb.ID, pb.Name)
	}
	// The whole document round-trips, not just the extracted columns.
	if len(got.Workflow) != 1 {
		t.Errorf("Get() workflow has %d steps, want 1", len(got.Workflow))
	}
}

func TestPlaybookCreateDuplicateIsConflict(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	pb := testPlaybook("playbook--dup")

	if err := store.Playbooks().Create(ctx, pb); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	err := store.Playbooks().Create(ctx, pb)
	if !errors.Is(err, storage.ErrConflict) {
		t.Errorf("Create() duplicate = %v, want ErrConflict", err)
	}
}

func TestPlaybookGetMissingIsNotFound(t *testing.T) {
	store := newTestStore(t)

	_, err := store.Playbooks().Get(context.Background(), "playbook--missing")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Get() missing = %v, want ErrNotFound", err)
	}
}

func TestPlaybookUpdate(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	pb := testPlaybook("playbook--upd")

	if err := store.Playbooks().Create(ctx, pb); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	pb.Name = "Renamed"
	if err := store.Playbooks().Update(ctx, pb); err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	got, err := store.Playbooks().Get(ctx, pb.ID)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if got.Name != "Renamed" {
		t.Errorf("Get() name = %q, want %q", got.Name, "Renamed")
	}
}

func TestPlaybookUpdateMissingIsNotFound(t *testing.T) {
	store := newTestStore(t)

	err := store.Playbooks().Update(context.Background(), testPlaybook("playbook--nope"))
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Update() missing = %v, want ErrNotFound", err)
	}
}

func TestPlaybookDelete(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	pb := testPlaybook("playbook--del")

	if err := store.Playbooks().Create(ctx, pb); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	if err := store.Playbooks().Delete(ctx, pb.ID); err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}
	if _, err := store.Playbooks().Get(ctx, pb.ID); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Get() after delete = %v, want ErrNotFound", err)
	}
	if err := store.Playbooks().Delete(ctx, pb.ID); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Delete() missing = %v, want ErrNotFound", err)
	}
}

func TestPlaybookListAndListMeta(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for _, id := range []string{"playbook--a", "playbook--b"} {
		if err := store.Playbooks().Create(ctx, testPlaybook(id)); err != nil {
			t.Fatalf("Create(%s) returned error: %v", id, err)
		}
	}

	list, err := store.Playbooks().List(ctx)
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("List() returned %d playbooks, want 2", len(list))
	}

	metas, err := store.Playbooks().ListMeta(ctx)
	if err != nil {
		t.Fatalf("ListMeta() returned error: %v", err)
	}
	if len(metas) != 2 {
		t.Fatalf("ListMeta() returned %d metas, want 2", len(metas))
	}
	if metas[0].ID != "playbook--a" {
		t.Errorf("ListMeta()[0].ID = %q, want playbook--a", metas[0].ID)
	}
	if len(metas[0].Labels) != 2 {
		t.Errorf("ListMeta()[0].Labels = %v, want 2 labels", metas[0].Labels)
	}
	if metas[0].ValidFrom.IsZero() {
		t.Error("ListMeta()[0].ValidFrom is zero, want the stored timestamp")
	}
}

func TestPlaybookListIsEmptyNotNil(t *testing.T) {
	store := newTestStore(t)

	list, err := store.Playbooks().List(context.Background())
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if list == nil {
		t.Error("List() returned nil, want empty slice")
	}
}

func testFin(id, tokenHash string) fin.Record {
	return fin.Record{
		FinId:           id,
		FinTokenHash:    tokenHash,
		DisplayName:     "test-fin",
		ProtocolVersion: "1",
		Capabilities:    []fin.Capability{{Type: "soarca-fin-test"}},
		RegisteredAt:    time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		LastSeen:        time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
	}
}

func TestFinCreateAndGet(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	record := testFin("fin--1", "hash-1")

	if err := store.Fins().Create(ctx, record); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	got, err := store.Fins().Get(ctx, record.FinId)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if got.DisplayName != record.DisplayName {
		t.Errorf("Get() display name = %q, want %q", got.DisplayName, record.DisplayName)
	}
	if len(got.Capabilities) != 1 || got.Capabilities[0].Type != "soarca-fin-test" {
		t.Errorf("Get() capabilities = %v, want one soarca-fin-test", got.Capabilities)
	}
}

func TestFinGetByTokenHash(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	record := testFin("fin--token", "hash-token")

	if err := store.Fins().Create(ctx, record); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	got, err := store.Fins().GetByTokenHash(ctx, "hash-token")
	if err != nil {
		t.Fatalf("GetByTokenHash() returned error: %v", err)
	}
	if got.FinId != record.FinId {
		t.Errorf("GetByTokenHash() fin id = %q, want %q", got.FinId, record.FinId)
	}

	if _, err := store.Fins().GetByTokenHash(ctx, "nope"); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("GetByTokenHash() unknown = %v, want ErrNotFound", err)
	}
}

func TestFinDuplicateTokenHashIsConflict(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.Fins().Create(ctx, testFin("fin--a", "shared-hash")); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	err := store.Fins().Create(ctx, testFin("fin--b", "shared-hash"))
	if !errors.Is(err, storage.ErrConflict) {
		t.Errorf("Create() duplicate token hash = %v, want ErrConflict", err)
	}
}

func TestFinTouch(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	record := testFin("fin--touch", "hash-touch")

	if err := store.Fins().Create(ctx, record); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	later := record.LastSeen.Add(time.Hour)
	if err := store.Fins().Touch(ctx, record.FinId, later); err != nil {
		t.Fatalf("Touch() returned error: %v", err)
	}

	got, err := store.Fins().Get(ctx, record.FinId)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if !got.LastSeen.UTC().Equal(later) {
		t.Errorf("LastSeen = %v, want %v", got.LastSeen.UTC(), later)
	}

	if err := store.Fins().Touch(ctx, "fin--missing", later); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Touch() missing = %v, want ErrNotFound", err)
	}
}

func TestFinListAndDelete(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.Fins().Create(ctx, testFin("fin--1", "h1")); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	if err := store.Fins().Create(ctx, testFin("fin--2", "h2")); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	records, err := store.Fins().List(ctx)
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("List() returned %d fins, want 2", len(records))
	}

	if err := store.Fins().Delete(ctx, "fin--1"); err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}
	if _, err := store.Fins().Get(ctx, "fin--1"); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Get() after delete = %v, want ErrNotFound", err)
	}
	if err := store.Fins().Delete(ctx, "fin--1"); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("Delete() missing = %v, want ErrNotFound", err)
	}
}

func TestMigrationsAreIdempotent(t *testing.T) {
	ctx := context.Background()
	url := "file:migrations-idempotent?mode=memory&cache=shared"

	first, err := New(ctx, url)
	if err != nil {
		t.Fatalf("first New() returned error: %v", err)
	}
	defer first.Close(ctx)

	second, err := New(ctx, url)
	if err != nil {
		t.Fatalf("second New() returned error: %v", err)
	}
	defer second.Close(ctx)
}

func TestParseURL(t *testing.T) {
	const fileDefaults = "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	const memoryDefaults = "?cache=shared&_pragma=busy_timeout(5000)"

	tests := []struct {
		url        string
		wantDriver string
		wantDSN    string
		wantPath   string
	}{
		{"sqlite://soarca.db", "sqlite", "file:soarca.db" + fileDefaults, "soarca.db"},
		{"sqlite://data/soarca.db", "sqlite", "file:data/soarca.db" + fileDefaults, "data/soarca.db"},
		{"sqlite:///var/lib/soarca.db", "sqlite", "file:/var/lib/soarca.db" + fileDefaults, "/var/lib/soarca.db"},
		{"sqlite3://soarca.db", "sqlite", "file:soarca.db" + fileDefaults, "soarca.db"},
		{"sqlite://:memory:", "sqlite", "file::memory:" + memoryDefaults, ""},
		{"./soarca.db", "sqlite", "file:./soarca.db" + fileDefaults, "./soarca.db"},
		// An explicit query takes over completely.
		{"sqlite://soarca.db?_pragma=journal_mode(DELETE)", "sqlite", "file:soarca.db?_pragma=journal_mode(DELETE)", "soarca.db"},
		// A raw file: DSN is never rewritten.
		{"file:test?mode=memory", "sqlite", "file:test?mode=memory", ""},
		{"postgres://u:p@host:5432/soarca", "pgx", "postgres://u:p@host:5432/soarca", ""},
		{"postgresql://u:p@host:5432/soarca", "pgx", "postgresql://u:p@host:5432/soarca", ""},
	}

	for _, test := range tests {
		got, err := parseURL(test.url)
		if err != nil {
			t.Errorf("parseURL(%q) returned error: %v", test.url, err)
			continue
		}
		if got.driverName != test.wantDriver || got.dsn != test.wantDSN || got.filePath != test.wantPath {
			t.Errorf("parseURL(%q) = %q/%q/%q, want %q/%q/%q",
				test.url, got.driverName, got.dsn, got.filePath,
				test.wantDriver, test.wantDSN, test.wantPath)
		}
	}
}

func TestParseURLRejectsUnknownScheme(t *testing.T) {
	if _, err := parseURL("mongodb://localhost:27017"); err == nil {
		t.Error("parseURL() accepted a mongodb URL, want an error")
	}
	if _, err := parseURL(""); err == nil {
		t.Error("parseURL() accepted an empty URL, want an error")
	}
}

func TestSQLiteFileCreatesParentDirectory(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "nested", "dir", "soarca.db")

	store, err := New(ctx, "sqlite://"+path)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer store.Close(ctx)

	if _, err := os.Stat(path); err != nil {
		t.Errorf("database file was not created: %v", err)
	}
}

// The pool is left at its defaults, so an in-memory database must survive
// queries landing on different connections.
func TestInMemoryStoreWorksAcrossPooledConnections(t *testing.T) {
	ctx := context.Background()

	store, err := New(ctx, "sqlite://:memory:")
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer store.Close(ctx)

	if err := store.Playbooks().Create(ctx, testPlaybook("playbook--pooled")); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	// Hold one connection open so the reads below need a second one.
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx() returned error: %v", err)
	}
	defer tx.Rollback()

	for i := 0; i < 5; i++ {
		if _, err := store.Playbooks().Get(ctx, "playbook--pooled"); err != nil {
			t.Fatalf("Get() on pooled connection returned error: %v", err)
		}
	}

	if open := store.db.Stats().OpenConnections; open < 2 {
		t.Errorf("only %d connection(s) opened, test did not exercise pooling", open)
	}
}

// A file database should end up in WAL mode so readers do not block a writer.
func TestFileStoreUsesWAL(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "soarca.db")

	store, err := New(ctx, "sqlite://"+path)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer store.Close(ctx)

	var mode string
	if err := store.db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatalf("PRAGMA journal_mode returned error: %v", err)
	}
	if !strings.EqualFold(mode, "wal") {
		t.Errorf("journal_mode = %q, want wal", mode)
	}
}
