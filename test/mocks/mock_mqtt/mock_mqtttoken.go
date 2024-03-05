package mock_mqtt

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type Mock_MqttToken struct {
	mock.Mock
}

func (token *Mock_MqttToken) Wait() bool {
	args := token.Called()
	return args.Bool(0)
}

func (token *Mock_MqttToken) WaitTimeout(duration time.Duration) bool {
	args := token.Called(duration)
	return args.Bool(0)
}

func (token *Mock_MqttToken) Done() <-chan struct{} {
	args := token.Called()
	return args.Get(0).(<-chan struct{})
}

func (token *Mock_MqttToken) Error() error {
	args := token.Called()
	return args.Error(0)
}
