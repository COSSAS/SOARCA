---
title:  Fin Python Library
description: >
    Documentation of the Python Fin library
categories: [extensions, architecture]
tags: [fin, python]
weight: 2
date: 2024-04-10
---

For the documentation about the Fin protocol we refer the reader to de documention page of [SOARCA Fin Protocol](https://cossas.github.io/SOARCA/docs/soarca-extensions/fin-protocol/)

## Quick Start
 
To include the SOARCA Fin library, you can use the following command to install it via pip:

```bash
pip install soarca-fin-library
```

### Example
An example on how to use the library is given below. 
For more examples and the source code, we will refer to the Github page of the [SOARCA-Fin-python-library](https://github.com/COSSAS/SOARCA-FIN-python-library).

```python
def capability_pong_callback(command: Command) -> ResultStructure:
    print("Received ping, returning pong!")

    result = Variable(
        type=VariableTypeEnum.string,
        name="pong_output",
        description="If ping, return pong",
        value="pong",
        constant=True,
        external=False)

    context = command.command.context

    return ResultStructure(
        state="success", context=context, variables={"result": result})


def main(mqtt_broker: str, mqtt_port: int, username: str, password: str) -> None:

    finId = "soarca-fin--pingpong-f877bb3a-bb37-429e-8ece-2d4286cf326d"
    agentName = "soarca-fin-pong-f896bb3b-bb37-429e-8ece-2d4286cf326d"
    externalReferenceName = "external-reference-example-name"
    capabilityId = "mod-pong--e896aa3b-bb37-429e-8ece-2d4286cf326d"

    # Create AgentStructure
    agent = AgentStructure(
        name=agentName)

    # Create ExternalReference
    external_reference = ExternalReference(name=externalReferenceName)

    # Create StepStructure
    step_structure = StepStructure(
        name="step_name",
        description="step description",
        external_references=[external_reference],
        command="pong",
        target=agentName)

    # Create CapabilityStructure
    capability_structure = CapabilityStructure(
        capability_id=capabilityId,
        type=WorkFlowStepEnum.action,
        name="Ping Pong capability",
        version="0.0.1",
        step={
            "test": step_structure},
        agent={
            "testagent": agent})

    # Create Soarca fin
    fin = SoarcaFin(finId)
    # Set config for MQTT Server
    fin.set_config_MQTT_server(mqtt_broker, mqtt_port, username, password)
    # Register Capabilities
    fin.create_fin_capability(capability_structure, capability_pong_callback)
    # Start the fin
    fin.start_fin()


if __name__ == "__main__":
    load_dotenv()
    MQTT_BROKER = os.getenv("MQTT_BROKER", "localhost")
    MQTT_PORT = int(os.getenv("MQTT_PORT", "1883"))
    USERNAME = os.getenv("MQTT_USERNAME", "soarca")
    PASSWD = os.getenv("MQTT_PASSWD", "password")

    main(MQTT_BROKER, MQTT_PORT, USERNAME, PASSWD)
```

## Architecture
The main object of the application is the `SoarcaFin` object, which is responsible for configuring and creating and controlling the capabilities.
The SoarcaFin creates `MQTTClient`s for each capability registered, plus one for registering, unregistering and controlling the fi itself.
`MQTTClient`s each have their own connection to the MQTT Broker and own `Parser` and `Executor` objects.
The `Parser` object parsers the raw MQTT messages and tries to convert them to one of the objects in `src/models`.
The `Executor` runs in their own thread and handles the actual execution of the messages.
The `Executor` polls a thread-safe queue for new messages and performs IO operations, such as sending messages to the MQTT broker and calling capability callbacks.

### Setup SOARCA Capabilities


To register a fin to SOARCA, first create a `SoarcaFin` object and pass the `fin_id` in the constructor. The SOARCA `fin_id` must be in the format of: `sourca-fin-<capability>-<uuid4>`.
Call `set_config_MQTT_server()` to set the required configurations for the fin to connect to the MQTT broker.
For each capability to be registered, call `create_fin_capability()`. The capability callback funtion should return an object of type `ResultStructure`.
When all capabilities are initialized, call `start_fin()` for the SOARCA Fin to connect to the MQTT broker and register itself to SOARCA.

An example is given in this project in the file [`examples/pong_example.py`]

### Class Overview
```plantuml
interface IParser {
 Message parse_on_message()
}

interface IMQTTClient {
 void on_connect()
 void on_message()
}

interface ISoarcaFin {
 void set_config_MQTTServer()
 void set_fin_capabilities()
 void start_fin()
}

interface IExecutor {
 void queue_message()
}


class SoarcaFin
class MQTTClient
class Parser
class Executor

ISoarcaFin <|.. SoarcaFin
IMQTTClient <|.. MQTTClient
IParser <|.. Parser
IExecutor <|.. Executor

IMQTTClient <- SoarcaFin
MQTTClient -> IExecutor
IParser <-MQTTClient
```

### Sequence Diagrams
#### Command
```plantuml
Soarca -> "MQTTClient (Capability 1)" : Command Message [Capability ID Topic]

"MQTTClient (Capability 1)" -> Parser : parse_on_message(message)
"MQTTClient (Capability 1)" <-- Parser : Message.Command

"MQTTClient (Capability 1)" -> "Executor (Capability 1)" : Command message
Soarca <-- "Executor (Capability 1)" : Ack

"Executor (Capability 1)" -> "Capability Callback" : Command
"Executor (Capability 1)" <-- "Capability Callback" : Result


Soarca <- "Executor (Capability 1)" : Result
Soarca --> "MQTTClient (Capability 1)" : Ack

"MQTTClient (Capability 1)" -> Parser : parse_on_message(message)
"MQTTClient (Capability 1)" <-- Parser : Message.Ack

"MQTTClient (Capability 1)" -> "Executor (Capability 1)" : Ack message
```

#### Register
```plantuml
Soarca -> Soarca : Create Soarca Topic

Library -> SoarcaFin : Set MQTT Server config

Library -> SoarcaFin : Set Capability1
SoarcaFin -> "MQTTClient (Capability 1)" : Create capability

Library -> SoarcaFin : Set Capability2
SoarcaFin -> "MQTTClient (Capability 2)" : Create capability


Library -> SoarcaFin : Start Fin


SoarcaFin -> "MQTTClient (Capability 1)" : Start capability
"MQTTClient (Capability 1)" -> "MQTTClient (Capability 1)" : Register Capability Topic
SoarcaFin -> "MQTTClient (Capability 2)" : Start capability
"MQTTClient (Capability 2)" -> "MQTTClient (Capability 2)" : Register Capability Topic

SoarcaFin -> "MQTTClient (Fin)" : Register Fin
"MQTTClient (Fin)" -> "MQTTClient (Fin)" : Register SoarcaFin Topic

"MQTTClient (Fin)" -> "Executor (Fin)" : Send Register Message

Soarca <- "Executor (Fin)" : Message.Register [Soarca Topic]

Soarca --> "MQTTClient (Fin)" :  Message.Ack [Fin ID Topic]

"MQTTClient (Fin)" -> "Parser (Fin)" : parse_on_message(ack)
"MQTTClient (Fin)" <-- "Parser (Fin)" : Message.Ack

"MQTTClient (Fin)" -> "Executor (Fin)" : Message.Ack
```

## Bugs or Contributing
Want to contribute to this project? It is possible to contribute [here](https://github.com/COSSAS/SOARCA-FIN-python-library).
Have you found a bug or want to request a feature? Please create an issue [here](https://github.com/COSSAS/SOARCA-FIN-python-library/issues).