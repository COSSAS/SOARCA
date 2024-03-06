package mock_decomposer

import (
	"soarca/internal/decomposer"
	"soarca/models/cacao"

	"github.com/stretchr/testify/mock"
)

type Mock_Decomposer struct {
	mock.Mock
}

func (mock *Mock_Decomposer) Execute(playbook cacao.Playbook) (*decomposer.ExecutionDetails, error) {
	args := mock.Called(playbook)
	return args.Get(0).(*decomposer.ExecutionDetails), args.Error(1)
}
