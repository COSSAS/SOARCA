package mock_walker

import (
	"soarca/internal/workflow"
	"soarca/pkg/models/cacao"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Walker struct {
	mock.Mock
}

func (mock *Mock_Walker) ExecuteAsync(playbook cacao.Playbook, results chan workflow.Result) {
	args := mock.Called(playbook, results)
	if results != nil {
		results <- workflow.Result{
			ExecutionId: args.Get(2).(uuid.UUID),
			PlaybookId:  playbook.ID,
			Variables:   cacao.NewVariables(),
		}
	}
}

func (mock *Mock_Walker) Execute(playbook cacao.Playbook) (*workflow.Result, error) {
	args := mock.Called(playbook)
	return args.Get(0).(*workflow.Result), args.Error(1)
}
