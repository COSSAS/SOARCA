package mock_playbook_database

import (
	"soarca/models/api"
	"soarca/models/cacao"

	"github.com/stretchr/testify/mock"
)

type MockPlaybook struct {
	mock.Mock
}

func (testInterface *MockPlaybook) GetPlaybookMetas() ([]api.PlaybookMeta, error) {
	args := testInterface.Called()
	return args.Get(0).([]api.PlaybookMeta), args.Error(1)
}

func (testInterface *MockPlaybook) GetPlaybooks() ([]cacao.Playbook, error) {
	args := testInterface.Called()
	return args.Get(0).([]cacao.Playbook), args.Error(1)
}

func (testInterface *MockPlaybook) Create(jsonData *[]byte) (cacao.Playbook, error) {
	args := testInterface.Called(jsonData)
	return args.Get(0).(cacao.Playbook), args.Error(1)
}

func (testInterface *MockPlaybook) Read(id string) (cacao.Playbook, error) {
	args := testInterface.Called(id)
	return args.Get(0).(cacao.Playbook), args.Error(1)
}

func (testInterface *MockPlaybook) Update(id string, jsonData *[]byte) (cacao.Playbook, error) {
	args := testInterface.Called(id, jsonData)
	return args.Get(0).(cacao.Playbook), args.Error(1)
}

func (testInterface *MockPlaybook) Delete(id string) error {
	args := testInterface.Called(id)
	return args.Error(0)
}
