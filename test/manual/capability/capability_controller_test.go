package capability_controller_test

import (
	"fmt"
	"soarca/pkg/core/capability/fin/controller"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func TestConnect(t *testing.T) {
	// used for manual testing

	// var executionId, _ = uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	// var playbookId = "playbook--d09351a2-a075-40c8-8054-0b7c423db83f"
	// var stepId = "step--81eff59f-d084-4324-9e0a-59e353dbd28f"
	// guid := new(guid.Guid)
	// prot := protocol.FinProtocol{Guid: guid, Topic: protocol.Topic("testing"), Broker: "localhost", Port: 1883}

	options := mqtt.NewClientOptions()
	options.AddBroker("mqtt://localhost:1883")
	options.SetClientID("soarca")
	options.SetUsername("public")
	options.SetPassword("password")

	client := mqtt.NewClient(options)

	finController := controller.New(client)

	if err := finController.ConnectAndSubscribe(); err != nil {
		fmt.Print(err)
		t.Fail()
	}
	finController.Run()

}
