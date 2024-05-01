package mock_time

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type MockTime struct {
	mock.Mock
}

func (t *MockTime) Now() time.Time {
	args := t.Called()
	return args.Get(0).(time.Time)
}
