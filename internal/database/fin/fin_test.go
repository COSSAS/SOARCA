package finrepository

import (
	"errors"
	"testing"
	"time"

	"soarca/pkg/models/fin"

	"github.com/go-playground/assert/v2"
)

// fakeDatabase is a minimal, in-memory database.Database test double used
// to exercise FinRepository's plumbing without needing a real MongoDB
// instance.
type fakeDatabase struct {
	records map[string]fin.Record
}

func newFakeDatabase() *fakeDatabase {
	return &fakeDatabase{records: map[string]fin.Record{}}
}

func (f *fakeDatabase) Read(id string) (any, error) {
	record, ok := f.records[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return record, nil
}

func (f *fakeDatabase) Find(query map[string]string, _ ...interface{}) ([]any, error) {
	results := []any{}
	for _, record := range f.records {
		if matches(record, query) {
			results = append(results, record)
		}
	}
	return results, nil
}

func matches(record fin.Record, query map[string]string) bool {
	for key, value := range query {
		switch key {
		case "fin_token_hash":
			if record.FinTokenHash != value {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func (f *fakeDatabase) Create(data interface{}) error {
	record, ok := data.(fin.Record)
	if !ok {
		return errors.New("unexpected type")
	}
	if _, exists := f.records[record.FinId]; exists {
		return errors.New("duplicate")
	}
	f.records[record.FinId] = record
	return nil
}

func (f *fakeDatabase) Update(id string, data interface{}) error {
	record, ok := f.records[id]
	if !ok {
		return errors.New("not found")
	}
	update, ok := data.(map[string]any)
	if !ok {
		return errors.New("unexpected type")
	}
	if lastSeen, ok := update["last_seen"].(time.Time); ok {
		record.LastSeen = lastSeen
	}
	f.records[id] = record
	return nil
}

func (f *fakeDatabase) Delete(id string) error {
	if _, ok := f.records[id]; !ok {
		return errors.New("not found")
	}
	delete(f.records, id)
	return nil
}

func TestRegisterAndGet(t *testing.T) {
	db := newFakeDatabase()
	repo := SetupFinRepository(db)

	record := fin.Record{
		FinId:        "fin-1",
		FinTokenHash: "hash-1",
		DisplayName:  "Example Pong Fin",
		Capabilities: []fin.Capability{{Type: "pong"}},
	}

	err := repo.Register(record)
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.Get("fin-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, got.FinId, "fin-1")
	assert.Equal(t, got.DisplayName, "Example Pong Fin")
	// RegisteredAt/LastSeen must be defaulted to "now" when not set by the
	// caller.
	if got.RegisteredAt.IsZero() {
		t.Fatal("expected RegisteredAt to be defaulted")
	}
	if got.LastSeen.IsZero() {
		t.Fatal("expected LastSeen to be defaulted")
	}
}

func TestRegisterDuplicateFinIdFails(t *testing.T) {
	db := newFakeDatabase()
	repo := SetupFinRepository(db)

	record := fin.Record{FinId: "fin-1", FinTokenHash: "hash-1"}
	if err := repo.Register(record); err != nil {
		t.Fatal(err)
	}
	if err := repo.Register(record); err == nil {
		t.Fatal("expected registering the same FinId twice to fail")
	}
}

func TestGetUnknownFinReturnsErrFinNotFound(t *testing.T) {
	db := newFakeDatabase()
	repo := SetupFinRepository(db)

	_, err := repo.Get("does-not-exist")
	var notFound fin.ErrFinNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected fin.ErrFinNotFound, got %v", err)
	}
	assert.Equal(t, notFound.FinId, "does-not-exist")
}

func TestFindByTokenHash(t *testing.T) {
	db := newFakeDatabase()
	repo := SetupFinRepository(db)

	record := fin.Record{FinId: "fin-1", FinTokenHash: "hash-1"}
	if err := repo.Register(record); err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindByTokenHash("hash-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, got.FinId, "fin-1")
}

func TestFindByTokenHashUnknownReturnsErrFinTokenInvalid(t *testing.T) {
	db := newFakeDatabase()
	repo := SetupFinRepository(db)

	_, err := repo.FindByTokenHash("does-not-exist")
	var invalid fin.ErrFinTokenInvalid
	if !errors.As(err, &invalid) {
		t.Fatalf("expected fin.ErrFinTokenInvalid, got %v", err)
	}
}

func TestList(t *testing.T) {
	db := newFakeDatabase()
	repo := SetupFinRepository(db)

	if err := repo.Register(fin.Record{FinId: "fin-1", FinTokenHash: "hash-1"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Register(fin.Record{FinId: "fin-2", FinTokenHash: "hash-2"}); err != nil {
		t.Fatal(err)
	}

	records, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(records), 2)
}

func TestTouchUpdatesLastSeen(t *testing.T) {
	db := newFakeDatabase()
	repo := SetupFinRepository(db)

	registeredAt := time.Now().Add(-time.Hour)
	if err := repo.Register(fin.Record{
		FinId:        "fin-1",
		FinTokenHash: "hash-1",
		RegisteredAt: registeredAt,
		LastSeen:     registeredAt,
	}); err != nil {
		t.Fatal(err)
	}

	newLastSeen := time.Now()
	if err := repo.Touch("fin-1", newLastSeen); err != nil {
		t.Fatal(err)
	}

	got, err := repo.Get("fin-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, got.LastSeen.Equal(newLastSeen), true)
	// RegisteredAt must be untouched by Touch.
	assert.Equal(t, got.RegisteredAt.Equal(registeredAt), true)
}

func TestUnregister(t *testing.T) {
	db := newFakeDatabase()
	repo := SetupFinRepository(db)

	if err := repo.Register(fin.Record{FinId: "fin-1", FinTokenHash: "hash-1"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Unregister("fin-1"); err != nil {
		t.Fatal(err)
	}

	_, err := repo.Get("fin-1")
	var notFound fin.ErrFinNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected fin.ErrFinNotFound after unregistering, got %v", err)
	}
}
