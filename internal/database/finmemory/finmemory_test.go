package finmemory

import (
	"errors"
	"testing"
	"time"

	"soarca/pkg/models/fin"

	"github.com/go-playground/assert/v2"
)

func TestRegisterAndGet(t *testing.T) {
	repo := New()

	record := fin.Record{FinId: "fin-1", FinTokenHash: "hash-1", DisplayName: "Example Pong Fin"}
	if err := repo.Register(record); err != nil {
		t.Fatal(err)
	}

	got, err := repo.Get("fin-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, got.DisplayName, "Example Pong Fin")
	if got.RegisteredAt.IsZero() || got.LastSeen.IsZero() {
		t.Fatal("expected RegisteredAt/LastSeen to be defaulted")
	}
}

func TestRegisterDuplicateFinIdFails(t *testing.T) {
	repo := New()
	record := fin.Record{FinId: "fin-1", FinTokenHash: "hash-1"}
	if err := repo.Register(record); err != nil {
		t.Fatal(err)
	}

	err := repo.Register(record)
	var alreadyRegistered fin.ErrAlreadyRegistered
	if !errors.As(err, &alreadyRegistered) {
		t.Fatalf("expected fin.ErrAlreadyRegistered, got %v", err)
	}
}

func TestGetUnknownFinReturnsErrFinNotFound(t *testing.T) {
	repo := New()
	_, err := repo.Get("does-not-exist")
	var notFound fin.ErrFinNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected fin.ErrFinNotFound, got %v", err)
	}
}

func TestFindByTokenHash(t *testing.T) {
	repo := New()
	if err := repo.Register(fin.Record{FinId: "fin-1", FinTokenHash: "hash-1"}); err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindByTokenHash("hash-1")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, got.FinId, "fin-1")

	_, err = repo.FindByTokenHash("does-not-exist")
	var invalid fin.ErrFinTokenInvalid
	if !errors.As(err, &invalid) {
		t.Fatalf("expected fin.ErrFinTokenInvalid, got %v", err)
	}
}

func TestList(t *testing.T) {
	repo := New()
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
	repo := New()
	if err := repo.Register(fin.Record{FinId: "fin-1", FinTokenHash: "hash-1"}); err != nil {
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
}

func TestTouchUnknownFinFails(t *testing.T) {
	repo := New()
	err := repo.Touch("does-not-exist", time.Now())
	var notFound fin.ErrFinNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected fin.ErrFinNotFound, got %v", err)
	}
}

func TestUnregister(t *testing.T) {
	repo := New()
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

func TestUnregisterUnknownFinFails(t *testing.T) {
	repo := New()
	err := repo.Unregister("does-not-exist")
	var notFound fin.ErrFinNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected fin.ErrFinNotFound, got %v", err)
	}
}
