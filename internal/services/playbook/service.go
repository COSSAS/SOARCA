package playbook

import (
	"context"

	"soarca/internal/storage"
	"soarca/pkg/models/api"
	"soarca/pkg/models/cacao"
)

// Service implements the services.PlaybookService interface.
type Service struct {
	store storage.PlaybookStore
}

// New creates a new playbook service.
func New(store storage.PlaybookStore) *Service {
	return &Service{store: store}
}

// ListPlaybooks returns all stored playbooks.
func (s *Service) ListPlaybooks(ctx context.Context) ([]cacao.Playbook, error) {
	return s.store.List(ctx)
}

// GetPlaybook retrieves a playbook by ID.
func (s *Service) GetPlaybook(ctx context.Context, playbookID string) (*cacao.Playbook, error) {
	playbook, err := s.store.Get(ctx, playbookID)
	if err != nil {
		return nil, err
	}
	return &playbook, nil
}

// CreatePlaybook stores a new playbook.
func (s *Service) CreatePlaybook(ctx context.Context, playbook *cacao.Playbook) error {
	return s.store.Create(ctx, *playbook)
}

// UpdatePlaybook replaces an existing playbook.
func (s *Service) UpdatePlaybook(ctx context.Context, playbookID string, playbook *cacao.Playbook) error {
	playbook.ID = playbookID
	return s.store.Update(ctx, *playbook)
}

// DeletePlaybook removes a playbook.
func (s *Service) DeletePlaybook(ctx context.Context, playbookID string) error {
	return s.store.Delete(ctx, playbookID)
}

// ListPlaybookMetas returns playbook metadata entries.
func (s *Service) ListPlaybookMetas(ctx context.Context) ([]api.PlaybookMeta, error) {
	return s.store.ListMeta(ctx)
}
