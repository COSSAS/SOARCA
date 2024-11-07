package controller_test

import (
	"encoding/json"
	"soarca/internal/capability/controller"
	"soarca/models/fin"
	"soarca/test/unittest/mocks/mock_mqtt"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestGetRegisteredc(t *testing.T) {
	mqtt := new(mock_mqtt.Mock_MqttClient)
	token := mock_mqtt.Mock_MqttToken{}
	token2 := mock_mqtt.Mock_MqttToken{}
	capabiltyController := controller.New(mqtt)
	fins := capabiltyController.GetRegisteredCapabilities()

	assert.Equal(t, len(fins), 0)

	messageId := uuid.New()

	capability := fin.Capability{Name: "cap1", Id: "id1", Version: "1.0.0"}
	capabilities := make([]fin.Capability, 0)
	capabilities = append(capabilities, capability)

	meta := fin.Meta{}

	incommingRegisterMessage := fin.Register{Type: fin.MessageTypeRegister,
		MessageId:       messageId.String(),
		FinID:           "Fin",
		ProtocolVersion: "1.0.0",
		Security:        fin.Security{Version: "0.0.0", ChannelSecurity: ""},
		Capabilities:    capabilities,
		Meta:            meta,
	}

	object, err := json.Marshal(incommingRegisterMessage)
	if err != nil {
		t.Fail()
	}

	token.On("Wait").Return(true)
	mqtt.On("Subscribe", "id1", uint8(1), mock.Anything).Return(&token)

	expectedAck := fin.NewAck(messageId.String())
	json, _ := fin.Encode(expectedAck)
	token2.On("Wait").Return(true)
	mqtt.On("Publish", "Fin", uint8(1), false, json).Return(&token2)
	token2.On("Error").Return(nil)

	capabiltyController.Handle(object)

	newFins := capabiltyController.GetRegisteredCapabilities()

	assert.Equal(t, len(newFins), 1)
	assert.Equal(t, newFins["id1"].Id, "id1")
	assert.Equal(t, newFins["id1"].Name, "cap1")
	mqtt.AssertExpectations(t)
	token.AssertExpectations(t)
	token2.AssertExpectations(t)

}

func TestConnectAndSubsribe(t *testing.T) {
	mqtt := new(mock_mqtt.Mock_MqttClient)
	token := mock_mqtt.Mock_MqttToken{}
	capabiltyController := controller.New(mqtt)

	token.On("Wait").Return(true)
	token.On("Error").Return(nil)
	mqtt.On("Connect").Return(&token)
	token.On("Wait").Return(true)
	token.On("Error").Return(nil)
	mqtt.On("Subscribe", "soarca", uint8(1), mock.Anything).Return(&token)
	err := capabiltyController.ConnectAndSubscribe()
	assert.Equal(t, err, nil)
	mqtt.AssertExpectations(t)
	token.AssertExpectations(t)
}
