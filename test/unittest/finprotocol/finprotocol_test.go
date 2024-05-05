package finprotocol_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"soarca/internal/fin/protocol"
	"soarca/internal/guid"
	"soarca/models/cacao"
	"soarca/models/fin"
	model "soarca/models/fin"
	"soarca/test/unittest/mocks/mock_mqtt"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/mock"
)

func TestSubscribe(t *testing.T) {
	mock_client := mock_mqtt.Mock_MqttClient{}
	mock_token := mock_mqtt.Mock_MqttToken{}

	guid := new(guid.Guid)
	prot := protocol.FinProtocol{Guid: guid, Topic: protocol.Topic("testing"), Broker: "localhost", Port: 1883}

	mock_token.On("Wait").Return(true)
	mock_client.On("Subscribe", "testing", uint8(1), mock.Anything).Return(&mock_token)
	prot.Subscribe(&mock_client)

}

func TestTimeoutAndCallbackTimerElaspsed(t *testing.T) {
	mock_client := mock_mqtt.Mock_MqttClient{}
	mock_token := mock_mqtt.Mock_MqttToken{}

	guid := new(guid.Guid)
	prot := protocol.FinProtocol{Guid: guid, Topic: protocol.Topic("testing"), Broker: "localhost", Port: 1883}

	mock_token.On("Wait").Return(true)
	mock_client.On("Subscribe", "testing", uint8(1), mock.Anything).Return(&mock_token)
	prot.Subscribe(&mock_client)

	expectedCommand := model.NewCommand()
	expectedCommand.CommandSubstructure.Context.Timeout = 1

	result, err := prot.AwaitResultOrTimeout(expectedCommand, &mock_client)

	assert.Equal(t, err, errors.New("no message received from fin while it was expected"))
	assert.Equal(t, result, map[string]cacao.Variable{})
}

func TestTimeoutAndCallbackHandlerCalled(t *testing.T) {
	mock_client := mock_mqtt.Mock_MqttClient{}
	mock_token := mock_mqtt.Mock_MqttToken{}

	mock_token_ack := mock_mqtt.Mock_MqttToken{}

	guid := new(guid.Guid)

	prot := protocol.New(guid, "testing", "localhost", 1883)
	mock_token.On("Wait").Return(true)
	mock_client.On("Subscribe", "testing", uint8(1), mock.Anything).Return(&mock_token)

	prot.Subscribe(&mock_client)

	expectedCommand := model.NewCommand()
	expectedCommand.CommandSubstructure.Context.Timeout = 1

	mock_token_ack.On("Wait").Return(true)
	mock_client.On("Publish", "testing", uint8(1), false, mock.Anything).Return(&mock_token_ack)

	fmt.Println("calling await")
	go helper(&prot)
	result, err := prot.AwaitResultOrTimeout(expectedCommand, &mock_client)
	fmt.Println("done waiting")

	assert.Equal(t, err, nil)
	assert.Equal(t, result, cacao.NewVariables(cacao.Variable{Name: "test"}))
	mock_client.AssertExpectations(t)
	mock_token.AssertExpectations(t)
	mock_token_ack.AssertExpectations(t)
}

// Helper for TestTimeoutAndCallbackHandlerCalled
func helper(prot *protocol.FinProtocol) {
	time.Sleep(1 * time.Millisecond)
	client := mock_mqtt.Mock_MqttClient{}
	message := mock_mqtt.Mock_MqttMessage{}

	ack := fin.Ack{}
	ack.Type = fin.MessageTypeAck
	ack.MessageId = "0001"
	ackPayload, err := json.Marshal(ack)
	if err != nil {
		fmt.Print(err)
		return
	}

	message.On("Topic").Return("testing")
	message.On("Payload").Return(ackPayload)
	fmt.Println("calling handler")
	prot.Handler(&client, &message)

	message2 := mock_mqtt.Mock_MqttMessage{}

	result := fin.Result{}
	result.Type = fin.MessageTypeResult
	result.ResultStructure.Variables = cacao.NewVariables(cacao.Variable{Name: "test"})

	payload, err := json.Marshal(result)
	if err != nil {
		fmt.Print(err)
		return
	}
	time.Sleep(1 * time.Millisecond)
	message2.On("Topic").Return("testing")
	message2.On("Payload").Return(payload)
	prot.Handler(&client, &message2)
	fmt.Println("called handler")

}
