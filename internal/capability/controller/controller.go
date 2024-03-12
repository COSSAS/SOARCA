package controller

import (
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/fin/protocol"
	"soarca/logger"
	"soarca/models/fin"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type CapabilityDetails struct {
	Name  string
	Id    string
	FinId string
}

const clientId = "soarca"

type IFinController interface {
	GetRegisteredCapabilities() map[string]CapabilityDetails
}

type FinController struct {
	registeredCapabilities map[string]CapabilityDetails
	mqttClient             mqtt.Client
	channel                chan []byte
}

func (finController *FinController) GetRegisteredCapabilities() map[string]CapabilityDetails {
	return finController.registeredCapabilities
}

func New(client mqtt.Client) *FinController {
	controllerQueue := make(chan []byte, 10)
	return &FinController{registeredCapabilities: make(map[string]CapabilityDetails), mqttClient: client, channel: controllerQueue}
}

func NewClient(url protocol.Broker, port int) *mqtt.Client {
	options := mqtt.NewClientOptions()
	options.AddBroker(fmt.Sprintf("mqtt://%s:%d", url, port))
	options.SetClientID(clientId)
	options.SetUsername("soarca")
	options.SetPassword("password")
	client := mqtt.NewClient(options)
	return &client
}

// This function will only return on a fatal error
func (finController *FinController) Start(broker string, port int) error {
	if finController.mqttClient == nil {
		return errors.New("fincontroller mqtt cilent is nil")
	}

	if token := finController.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		err := token.Error()
		log.Error(err)
		return err
	}

	token := finController.mqttClient.Subscribe(string("soarca"), 1, finController.Handler)
	token.Wait()

	for {
		select {
		case result := <-finController.channel:
			finController.Handle(result)
		}

	}

	return nil
}

// Handle goroutine call from mqtt stack
func (finController *FinController) Handler(client mqtt.Client, msg mqtt.Message) {
	if msg.Topic() != string("soarca") {
		log.Trace("message was receive in wrong topic: " + msg.Topic())
	}
	payload := msg.Payload()
	log.Trace(string(payload))
	finController.channel <- payload

}

func (finController *FinController) SendAck(topic string, messageId string) error {
	json, _ := fin.Encode(fin.NewAck(messageId))
	log.Trace("Sending ack for message id: ", messageId)
	token := finController.mqttClient.Publish(topic, 1, false, json)
	token.Wait()
	if err := token.Error(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (finController *FinController) Handle(payload []byte) {
	message := fin.Message{}
	fin.Decode(payload, &message)
	switch message.Type {
	case fin.MessageTypeAck:
		finController.HandleAck(payload)
	case fin.MessageTypeRegister:
		finController.HandleRegister(payload)
	case fin.MessageTypeNack:
		finController.HandleNack(payload)
	}
}

func (finController *FinController) SendNack(topic string, messageId string) error {
	json, _ := fin.Encode(fin.NewNack(messageId))
	log.Trace("Sending nack for message id: ", messageId)
	token := finController.mqttClient.Publish(topic, 1, false, json)
	token.Wait()
	if err := token.Error(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (finController *FinController) HandleAck(payload []byte) {
	ack := fin.Ack{}
	fin.Decode(payload, ack)

	// ignore for now

}

func (finController *FinController) HandleNack(payload []byte) {
	ack := fin.Ack{}
	fin.Decode(payload, ack)

	// ignore for now

}

func (finController *FinController) HandleRegister(payload []byte) {
	register := fin.Register{}
	err := fin.Decode(payload, &register)
	if err != nil {
		log.Error("Message", err)
		finController.SendNack("soarca", register.MessageId)
		return
	}

	for _, capability := range register.Capabilities {
		if _, ok := finController.registeredCapabilities[capability.Id]; ok {
			finController.SendNack(register.FinID, register.MessageId)
			log.Error("this capability UUID is already registered")
			return
		}
		token := finController.mqttClient.Subscribe(capability.Id, 1, finController.Handler)
		token.Wait()

		detail := CapabilityDetails{Name: capability.Name, Id: capability.Id, FinId: register.FinID}
		finController.registeredCapabilities[capability.Id] = detail

	}

	finController.SendAck(register.FinID, register.MessageId)

}
