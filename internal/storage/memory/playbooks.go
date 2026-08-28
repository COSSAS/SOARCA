package memory

import (
	"context"
	"sync"

	"soarca/internal/storage"
	"soarca/pkg/models/api"
	"soarca/pkg/models/cacao"
)

var _ storage.PlaybookStore = (*PlaybookStore)(nil)

type PlaybookStore struct {
	mu        sync.RWMutex
	playbooks map[string]cacao.Playbook
}

func newPlaybookStore() *PlaybookStore {
	return &PlaybookStore{playbooks: make(map[string]cacao.Playbook)}
}

func (s *PlaybookStore) Create(_ context.Context, pb cacao.Playbook) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.playbooks[pb.ID]; exists {
		return storage.ErrConflict
	}
	s.playbooks[pb.ID] = pb
	return nil
}

func (s *PlaybookStore) Update(_ context.Context, pb cacao.Playbook) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.playbooks[pb.ID]; !exists {
		return storage.ErrNotFound
	}
	s.playbooks[pb.ID] = pb
	return nil
}

func (s *PlaybookStore) Get(_ context.Context, id string) (cacao.Playbook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pb, ok := s.playbooks[id]
	if !ok {
		return cacao.Playbook{}, storage.ErrNotFound
	}
	return pb, nil
}

func (s *PlaybookStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.playbooks[id]; !ok {
		return storage.ErrNotFound
	}
	delete(s.playbooks, id)
	return nil
}

func (s *PlaybookStore) List(_ context.Context) ([]cacao.Playbook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]cacao.Playbook, 0, len(s.playbooks))
	for _, pb := range s.playbooks {
		result = append(result, pb)
	}
	return result, nil
}

func (s *PlaybookStore) ListMeta(_ context.Context) ([]api.PlaybookMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]api.PlaybookMeta, 0, len(s.playbooks))
	for _, pb := range s.playbooks {
		result = append(result, api.PlaybookMeta{
			ID:          pb.ID,
			Name:        pb.Name,
			Description: pb.Description,
			ValidFrom:   pb.ValidFrom,
			ValidUntil:  pb.ValidUntil,
			Labels:      pb.Labels,
		})
	}
	return result, nil
}
