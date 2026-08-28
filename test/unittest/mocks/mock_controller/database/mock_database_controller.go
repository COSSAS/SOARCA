package mock_database_controller

import (
	"soarca/internal/storage"

	"github.com/stretchr/testify/mock"
)

type Mock_Controller struct {
	mock.Mock
}

func (mock *Mock_Controller) GetPlaybookStore() storage.PlaybookStore {
	args := mock.Called()
	return args.Get(0).(storage.PlaybookStore)
}
