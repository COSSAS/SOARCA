package mock_finprotocol

import (
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/fin"

	"github.com/stretchr/testify/mock"
)

type MockFinProtocol struct {
	mock.Mock
}

func (finProtocol *MockFinProtocol) SendCommand(command fin.Command) (cacao.Variables, error) {
	args := finProtocol.Called(command)
	return args.Get(0).(cacao.Variables), args.Error(1)
}
