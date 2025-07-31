package mock_guid

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock_Guid struct {
	mock.Mock
}

func (guid *Mock_Guid) New() uuid.UUID {
	args := guid.Called()
	return args.Get(0).(uuid.UUID)
}

func (guid *Mock_Guid) NewV7() uuid.UUID {
	args := guid.Called()
	return args.Get(0).(uuid.UUID)
}
