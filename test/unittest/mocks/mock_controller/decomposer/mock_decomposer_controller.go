package mock_decomposer_controller

import (
	"soarca/internal/decomposer"

	"github.com/stretchr/testify/mock"
)

type Mock_Controller struct {
	mock.Mock
}

func (mock *Mock_Controller) NewDecomposer() decomposer.IDecomposer {
	args := mock.Called()
	return args.Get(0).(decomposer.IDecomposer)
}
