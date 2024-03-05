package mock_mqtt

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/stretchr/testify/mock"
)

type Mock_MqttClient struct {
	mock.Mock
}

func (client *Mock_MqttClient) IsConnected() bool {
	args := client.Called()
	return args.Get(0).(bool)
}

func (client *Mock_MqttClient) IsConnectionOpen() bool {
	args := client.Called()
	return args.Get(0).(bool)
}

func (client *Mock_MqttClient) Connect() mqtt.Token {
	args := client.Called()
	return args.Get(0).(mqtt.Token)
}

func (client *Mock_MqttClient) Disconnect(quiesce uint) {
	client.Called(quiesce)
}

func (client *Mock_MqttClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqtt.Token {
	args := client.Called(topic, qos, retained, payload)
	return args.Get(0).(mqtt.Token)
}

func (client *Mock_MqttClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
	args := client.Called(topic, qos, callback)
	return args.Get(0).(mqtt.Token)
}

func (client *Mock_MqttClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	args := client.Called(filters, callback)
	return args.Get(0).(mqtt.Token)
}

func (client *Mock_MqttClient) AddRoute(topic string, callback mqtt.MessageHandler) {
	client.Called(topic, callback)
}

func (client *Mock_MqttClient) OptionsReader() mqtt.ClientOptionsReader {
	args := client.Called()
	return args.Get(0).(mqtt.ClientOptionsReader)
}

func (client *Mock_MqttClient) Unsubscribe(topics ...string) mqtt.Token {
	args := client.Called(topics)
	return args.Get(0).(mqtt.Token)
}
