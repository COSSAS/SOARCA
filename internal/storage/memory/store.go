package memory

import (
	"context"

	"soarca/internal/storage"
)

var _ storage.Store = (*Store)(nil)

type Store struct {
	playbooks *PlaybookStore
	fins      *FinStore
}

func New() *Store {
	return &Store{
		playbooks: newPlaybookStore(),
		fins:      newFinStore(),
	}
}

func (s *Store) Playbooks() storage.PlaybookStore { return s.playbooks }
func (s *Store) Fins() storage.FinStore           { return s.fins }
func (s *Store) Close(_ context.Context) error    { return nil }
