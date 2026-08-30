package mock_capability

import (
	"soarca/internal/workflow/capability"
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"

	"github.com/stretchr/testify/mock"
)

type Mock_Capability struct {
	mock.Mock
}

func (capability *Mock_Capability) Execute(metadata run.Metadata,
	context capability.Context) (cacao.Variables, error) {
	args := capability.Called(metadata, context)
	return args.Get(0).(cacao.Variables), args.Error(1)
}

func (capability *Mock_Capability) GetType() string {
	args := capability.Called()
	return args.Get(0).(string)
}
