package mock_assignment_extension

import (
	"soarca/pkg/extensions/soarca/assignment"
	"soarca/pkg/models/cacao"

	"github.com/stretchr/testify/mock"
)

type Mock_AssignmentExtension struct {
	mock.Mock
}

func (mock *Mock_AssignmentExtension) AssignAndEvaluate(assignment.Context) cacao.Variables {
	args := mock.Called()
	return args.Get(0).(cacao.Variables)
}
