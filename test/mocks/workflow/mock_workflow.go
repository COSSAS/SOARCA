package mocks_workflow_test

import (
	"soarca/models/api"
	"soarca/models/cacao"

	"github.com/stretchr/testify/mock"
)

type MockWorkflow struct {
	mock.Mock
}

func (testInterface *MockWorkflow) GetWorkflowMetas() ([]api.PlaybookMeta, error) {
	args := testInterface.Called()
	return args.Get(0).([]api.PlaybookMeta), args.Error(1)
}

func (testInterface *MockWorkflow) GetWorkflows() ([]cacao.Playbook, error) {
	args := testInterface.Called()
	return args.Get(0).([]cacao.Playbook), args.Error(1)
}

func (testInterface *MockWorkflow) Create(jsonData *[]byte) (cacao.Playbook, error) {
	args := testInterface.Called(jsonData)
	return args.Get(0).(cacao.Playbook), args.Error(1)
}

func (testInterface *MockWorkflow) Read(id string) (cacao.Playbook, error) {
	args := testInterface.Called(id)
	return args.Get(0).(cacao.Playbook), args.Error(1)
}

func (testInterface *MockWorkflow) Update(id string, jsonData *[]byte) (cacao.Playbook, error) {
	args := testInterface.Called(id, jsonData)
	return args.Get(0).(cacao.Playbook), args.Error(1)
}

func (testInterface *MockWorkflow) Delete(id string) error {
	args := testInterface.Called(id)
	return args.Error(0)
}
