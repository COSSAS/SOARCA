package mock_database_controller

import (
	playbookrepository "soarca/database/playbook"

	"github.com/stretchr/testify/mock"
)

type Mock_Controller struct {
	mock.Mock
}

func (mock *Mock_Controller) GetDatabaseInstance() playbookrepository.IPlaybookRepository {
	args := mock.Called()
	return args.Get(0).(playbookrepository.IPlaybookRepository)
}
