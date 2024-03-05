package mock_mqtt

import (
	"github.com/stretchr/testify/mock"
)

type Message interface {
	Duplicate() bool
	Qos() byte
	Retained() bool
	Topic() string
	MessageID() uint16
	Payload() []byte
	Ack()
}

type Mock_MqttMessage struct {
	mock.Mock
}

func (message *Mock_MqttMessage) Duplicate() bool {
	args := message.Called()
	return args.Bool(0)
}

func (message *Mock_MqttMessage) Qos() byte {
	args := message.Called()
	return args.Get(0).(byte)
}

func (message *Mock_MqttMessage) Retained() bool {
	args := message.Called()
	return args.Bool(0)
}

func (message *Mock_MqttMessage) Topic() string {
	args := message.Called()
	return args.String(0)
}

func (message *Mock_MqttMessage) MessageID() uint16 {
	args := message.Called()
	return args.Get(0).(uint16)
}

func (message *Mock_MqttMessage) Payload() []byte {
	args := message.Called()
	return args.Get(0).([]byte)
}

func (message *Mock_MqttMessage) Ack() {
	message.Called()
}
