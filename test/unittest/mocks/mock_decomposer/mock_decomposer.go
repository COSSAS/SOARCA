package mock_decomposer

import (
	"soarca/internal/decomposer"
	"soarca/models/cacao"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Decomposer struct {
	mock.Mock
}

func (mock *Mock_Decomposer) ExecuteAsync(playbook cacao.Playbook, detailsch chan decomposer.ExecutionDetails) {
	args := mock.Called(playbook, detailsch)
	if detailsch != nil {
		details := decomposer.ExecutionDetails{ExecutionId: args.Get(2).(uuid.UUID), PlaybookId: playbook.ID, Variables: cacao.NewVariables()}
		detailsch <- details
	}
}
func (mock *Mock_Decomposer) Execute(playbook cacao.Playbook) (*decomposer.ExecutionDetails, error) {
	args := mock.Called(playbook)
	return args.Get(0).(*decomposer.ExecutionDetails), args.Error(1)
}
