package protocol

import (
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/fin"
	"soarca/pkg/utils/guid"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

const defaultTimeout = 1
const disconnectTimeout = 100
const clientId = "soarca-fin-capability"
const defaultQos = AtLeastOnce

const (
	AtMostOnce = iota
	AtLeastOnce
	ExactlyOnce
)

type Topic string
type Message string
type Broker string

var component = reflect.TypeOf(FinProtocol{}).PkgPath()
var log *logger.Log

// var channel = make(chan []byte, 1)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type IFinProtocol interface {
	SendCommand(fin.Command) (cacao.Variables, error)
}

type FinProtocol struct {
	Topic   Topic
	Broker  Broker
	Port    int
	Guid    guid.IGuid
	channel chan []byte // Channel is for one instance and is private to this fin
}

func New(guid guid.IGuid, topic Topic, broker Broker, port int) FinProtocol {
	var channel = make(chan []byte, 1)
	prot := FinProtocol{Guid: guid, Topic: topic, Broker: broker, Port: port, channel: channel}
	return prot
}

func (protocol *FinProtocol) SendAck(result fin.Result, client mqttlib.Client) {
	ack := fin.NewAck(result.MessageId)
	json, _ := fin.Encode(ack)
	log.Trace("Sending ack for message id: ", result.MessageId)
	token := client.Publish(string(protocol.Topic), defaultQos, false, json)
	token.Wait()
}

func (protocol *FinProtocol) SendNack(result fin.Result, client mqttlib.Client) {
	nack := fin.NewNack(result.MessageId)
	json, _ := fin.Encode(nack)
	log.Trace("Sending nack for message id: ", result.MessageId)
	token := client.Publish(string(protocol.Topic), defaultQos, false, json)
	token.Wait()
}

func (protocol *FinProtocol) SendCommand(command fin.Command) (cacao.Variables, error) {

	client, err := protocol.Connect(command.CommandSubstructure.Authentication)
	if err != nil {
		log.Error("could not connect to mqtt broker")
		return nil, err
	}

	protocol.Subscribe(client)
	err = protocol.Publish(client, command)
	if err != nil {
		protocol.Disconnect(client)
		return cacao.NewVariables(), err
	}
	result, err := protocol.AwaitResultOrTimeout(command, client)
	protocol.Disconnect(client)

	return result, err
}

func (protocol *FinProtocol) AwaitResultOrTimeout(command fin.Command, client mqttlib.Client) (cacao.Variables, error) {
	timeout := command.CommandSubstructure.Context.Timeout

	if command.CommandSubstructure.Context.Timeout == 0 {
		log.Warning("no valid timeout will set 1 second")
		timeout = defaultTimeout
	}
	timer := time.NewTimer(time.Duration(timeout) * time.Second)

	// Wait in a loop for the timer to elapse or a message on the channel
	ackReceived := false

	for {
		select {
		case <-timer.C:
			err := errors.New("no message received from fin while it was expected")
			return cacao.NewVariables(), err
		case result := <-protocol.channel:
			finMessage := fin.Message{}
			err := fin.Decode(result, &finMessage)
			if err != nil {
				log.Trace(err)
				break
			}
			log.Info(finMessage)

			// This now accepts any ack, should be changed
			switch finMessage.Type {
			case fin.MessageTypeAck:
				ackReceived = true
			case fin.MessageTypeResult:
				finResult := fin.Result{}
				err := fin.Decode(result, &finResult)
				if err != nil {
					log.Trace(err)
					return cacao.NewVariables(), err
				}

				if ackReceived {

					if finResult.ResultStructure.Context.ExecutionId == command.CommandSubstructure.Context.ExecutionId {

						protocol.SendAck(finResult, client)
						return finResult.ResultStructure.Variables, nil
					} else {
						protocol.SendNack(finResult, client)
					}

				} else {
					protocol.SendNack(finResult, client)
				}
			}
		}

	}

}

func (protocol *FinProtocol) Handler(client mqttlib.Client, msg mqttlib.Message) {
	if msg.Topic() != string(protocol.Topic) {
		log.Trace("message was receive in wrong topic: " + protocol.Topic)
	}
	payload := msg.Payload()
	log.Trace(string(payload))
	protocol.channel <- payload

}

func (protocol *FinProtocol) Subscribe(client mqttlib.Client) {
	token := client.Subscribe(string(protocol.Topic), defaultQos, protocol.Handler)
	token.Wait()

}

func (protocol *FinProtocol) Publish(client mqttlib.Client, command fin.Command) error {
	command.MessageId = protocol.Guid.New().String()
	command.Meta.SenderId = clientId
	command.Meta.Timestamp = time.Now()

	data, err := fin.Encode(command)
	if err != nil {
		return err
	}
	token := client.Publish(string(protocol.Topic), defaultQos, false, data)
	token.Wait()
	return token.Error()

}

func (protocol *FinProtocol) Connect(authenticationInformation cacao.AuthenticationInformation) (mqttlib.Client, error) {
	options := mqttlib.NewClientOptions()
	options.AddBroker(fmt.Sprintf("mqtt://%s:%d", protocol.Broker, protocol.Port))
	options.SetClientID(clientId)
	options.SetUsername(authenticationInformation.Username)
	options.SetPassword(authenticationInformation.Password)

	client := mqttlib.NewClient(options)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		err := token.Error()
		log.Error(err)
		return nil, err
	}
	return client, nil
}

func (protocol *FinProtocol) Disconnect(client mqttlib.Client) {
	client.Disconnect(disconnectTimeout)
}
