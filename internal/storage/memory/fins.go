package memory

import (
	"context"
	"sync"
	"time"

	"soarca/internal/storage"
	"soarca/pkg/models/fin"
)

var _ storage.FinStore = (*FinStore)(nil)

type FinStore struct {
	mu      sync.RWMutex
	records map[string]fin.Record
}

func newFinStore() *FinStore {
	return &FinStore{records: make(map[string]fin.Record)}
}

func (s *FinStore) Create(_ context.Context, record fin.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.records[record.FinId]; exists {
		return storage.ErrConflict
	}
	s.records[record.FinId] = record
	return nil
}

func (s *FinStore) Get(_ context.Context, finID string) (fin.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[finID]
	if !ok {
		return fin.Record{}, storage.ErrNotFound
	}
	return record, nil
}

func (s *FinStore) GetByTokenHash(_ context.Context, tokenHash string) (fin.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, record := range s.records {
		if record.FinTokenHash == tokenHash {
			return record, nil
		}
	}
	return fin.Record{}, storage.ErrNotFound
}

func (s *FinStore) List(_ context.Context) ([]fin.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]fin.Record, 0, len(s.records))
	for _, record := range s.records {
		result = append(result, record)
	}
	return result, nil
}

func (s *FinStore) Touch(_ context.Context, finID string, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.records[finID]
	if !ok {
		return storage.ErrNotFound
	}
	record.LastSeen = at
	s.records[finID] = record
	return nil
}

func (s *FinStore) Delete(_ context.Context, finID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.records[finID]; !ok {
		return storage.ErrNotFound
	}
	delete(s.records, finID)
	return nil
}
