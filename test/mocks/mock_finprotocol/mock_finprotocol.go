package mock_finprotocol

import (
	"soarca/models/cacao"
	"soarca/models/fin"

	"github.com/stretchr/testify/mock"
)

type MockFinProtocol struct {
	mock.Mock
}

func (finProtocol *MockFinProtocol) SendCommand(command fin.Command) (map[string]cacao.Variable, error) {
	args := finProtocol.Called(command)
	return args.Get(0).(map[string]cacao.Variable), args.Error(1)
}
