package mock_database_controller

import (
	"context"
	"soarca/internal/store"
	"soarca/internal/transport/http/schema"
	"soarca/pkg/cacao"

	"github.com/stretchr/testify/mock"
)

type MockPlaybook struct {
	mock.Mock
}

var _ storage.PlaybookStore = (*MockPlaybook)(nil)

// New storage interface
func (m *MockPlaybook) Create(ctx context.Context, pb cacao.Playbook) error {
	args := m.Called(ctx, pb)
	return args.Error(0)
}

func (m *MockPlaybook) Update(ctx context.Context, pb cacao.Playbook) error {
	args := m.Called(ctx, pb)
	return args.Error(0)
}

func (m *MockPlaybook) Get(ctx context.Context, id string) (cacao.Playbook, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(cacao.Playbook), args.Error(1)
}

func (m *MockPlaybook) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPlaybook) List(ctx context.Context) ([]cacao.Playbook, error) {
	args := m.Called(ctx)
	return args.Get(0).([]cacao.Playbook), args.Error(1)
}

func (m *MockPlaybook) ListMeta(ctx context.Context) ([]api.PlaybookMeta, error) {
	args := m.Called(ctx)
	return args.Get(0).([]api.PlaybookMeta), args.Error(1)
}

// Legacy repository methods kept for older tests that still call them.
func (m *MockPlaybook) GetPlaybookMetas() ([]api.PlaybookMeta, error) { return m.ListMeta(context.Background()) }
func (m *MockPlaybook) GetPlaybooks() ([]cacao.Playbook, error)       { return m.List(context.Background()) }
func (m *MockPlaybook) Read(id string) (cacao.Playbook, error)        { return m.Get(context.Background(), id) }
func (m *MockPlaybook) CreateJSON(jsonData *[]byte) (cacao.Playbook, error) {
	args := m.Called(jsonData)
	return args.Get(0).(cacao.Playbook), args.Error(1)
}
func (m *MockPlaybook) UpdateJSON(id string, jsonData *[]byte) (cacao.Playbook, error) {
	args := m.Called(id, jsonData)
	return args.Get(0).(cacao.Playbook), args.Error(1)
}
func (m *MockPlaybook) DeleteLegacy(id string) error { return m.Delete(context.Background(), id) }
