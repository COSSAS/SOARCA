package mock_decomposer

import (
	"soarca/internal/decomposer"
	"soarca/models/cacao"

	"github.com/stretchr/testify/mock"
)

type Mock_Decomposer struct {
	mock.Mock
}

func (mock *Mock_Decomposer) Execute(playbook cacao.Playbook, detailsch chan string) (*decomposer.ExecutionDetails, error) {
	args := mock.Called(playbook, detailsch)
	if detailsch != nil {
		execution_ids := playbook.ID + "///" + "mock_uuid_string_defined_in_mock_decomposer"
		detailsch <- execution_ids
	}
	return args.Get(0).(*decomposer.ExecutionDetails), args.Error(1)
}
