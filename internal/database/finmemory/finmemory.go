// Package finmemory is a non-persisted fallback implementation of
// finrepository.IFinRepository, used when SOARCA is running without a
// configured database (mirroring internal/database/memory's fallback for
// playbooks). Fin registrations made against this implementation do not
// survive a SOARCA restart - the whole point of persisting registrations is
// defeated without a real database, so this is only intended for local/dev
// use, matching the playbook fallback's existing caveat.
package finmemory

import (
	"sync"
	"time"

	finrepository "soarca/internal/database/fin"
	"soarca/pkg/models/fin"
)

var _ finrepository.IFinRepository = (*InMemoryFinRepository)(nil)

type InMemoryFinRepository struct {
	mu      sync.Mutex
	records map[string]fin.Record
}

func New() *InMemoryFinRepository {
	return &InMemoryFinRepository{records: map[string]fin.Record{}}
}

func (memory *InMemoryFinRepository) Register(record fin.Record) error {
	memory.mu.Lock()
	defer memory.mu.Unlock()

	if _, exists := memory.records[record.FinId]; exists {
		return fin.ErrAlreadyRegistered{FinId: record.FinId}
	}
	if record.RegisteredAt.IsZero() {
		record.RegisteredAt = time.Now()
	}
	if record.LastSeen.IsZero() {
		record.LastSeen = record.RegisteredAt
	}
	memory.records[record.FinId] = record
	return nil
}

func (memory *InMemoryFinRepository) Get(finId string) (fin.Record, error) {
	memory.mu.Lock()
	defer memory.mu.Unlock()

	record, ok := memory.records[finId]
	if !ok {
		return fin.Record{}, fin.ErrFinNotFound{FinId: finId}
	}
	return record, nil
}

func (memory *InMemoryFinRepository) FindByTokenHash(tokenHash string) (fin.Record, error) {
	memory.mu.Lock()
	defer memory.mu.Unlock()

	for _, record := range memory.records {
		if record.FinTokenHash == tokenHash {
			return record, nil
		}
	}
	return fin.Record{}, fin.ErrFinTokenInvalid{}
}

func (memory *InMemoryFinRepository) List() ([]fin.Record, error) {
	memory.mu.Lock()
	defer memory.mu.Unlock()

	records := make([]fin.Record, 0, len(memory.records))
	for _, record := range memory.records {
		records = append(records, record)
	}
	return records, nil
}

func (memory *InMemoryFinRepository) Touch(finId string, lastSeen time.Time) error {
	memory.mu.Lock()
	defer memory.mu.Unlock()

	record, ok := memory.records[finId]
	if !ok {
		return fin.ErrFinNotFound{FinId: finId}
	}
	record.LastSeen = lastSeen
	memory.records[finId] = record
	return nil
}

func (memory *InMemoryFinRepository) Unregister(finId string) error {
	memory.mu.Lock()
	defer memory.mu.Unlock()

	if _, ok := memory.records[finId]; !ok {
		return fin.ErrFinNotFound{FinId: finId}
	}
	delete(memory.records, finId)
	return nil
}
