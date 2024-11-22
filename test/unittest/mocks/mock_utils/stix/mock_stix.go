package mock_stix

import (
	"soarca/pkg/models/cacao"

	"github.com/stretchr/testify/mock"
)

type MockStix struct {
	mock.Mock
}

type MockHttpRequest struct {
	mock.Mock
}

func (stix *MockStix) Evaluate(expression string, vars cacao.Variables) (bool, error) {
	args := stix.Called(expression, vars)
	return args.Bool(0), args.Error(1)
}
