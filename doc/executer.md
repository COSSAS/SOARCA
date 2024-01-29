# SOARCA Executer design

The document contains the design considerations of the executer of SOARCA

## Components

The executer consists of the following components. 

. The capability selector
. Native capabilities (command executors)
. MQTT capability to interact with: Fin capabilities (third party executors)

### Capability selector (Executor)

The capability selector will select the implementation which is capable of executing the incoming command. There are native capabilities which are based on the CACAO `command-type-ov`:

* Currently implemented:
    * ssh
    * http-api
* Later:
    * open-C2
    * manual
* Future:
    * bash
    * caldera-cmd
    * elastic
    * jupyter
    * kestrel
    * sigma
    * yara

### Native capabilities
The Executor will select a module which is capable of execution the command and pass the detail to it. The results will be returned to the decomposer. Result can be output variables or error status.

### MQTT executor -> Fin capabilities
The Executor will put the command on the MQTT topic that is offered by the module. How a module handles this is described in the link:modules.adoc[module documentation]

### Component overview

```plantuml

package "Controller" {
component Decomposer as parser

}
package "Executor" {
    component SSH as exe2
    component "HTTP-API" as exe1
    component MQTT as exe3
}

package "Fins" {
    component "Virus Total" as virustotal
    component "E-Sender" as email
}

parser -- Executor
exe3 -- Fins : " MQTT topics"
```


## Executor classes


```plantuml

interface IExecutor {
    void Execute(ExecutionId, CommandData, variable[], target, module, completionCallback(variables[]))
    void Pause(CommandData, module)
    void Resume(CommandData, module)
    void Kill(CommandData, module)
}

struct State{
    stopped
    running
    paused
}


interface ICapability{
    void Execute(ExecutionId, CommandData, variable[], target, completionCallback(variables[]))
    module GetModuleName()
}


class Executor 

class "Ssh" as ssh
class "OpenC2" as openc2
class "HttpApi" as api


IExecutor <|.. Executor
Executor -> ICapability
ICapability <|.. ssh
ICapability <|.. openc2
ICapability <|.. api

```




## Protocol (WIP)
https://github.com/phf/go-queue

### Sending a step

Variables can be input and output the CACAO spec should declare the following fields for variables:
direction : input|output


And every step that will output any variables should declare a list of output variables:
output: var_1

```plantuml
@startjson
{
        "execute-id" : "uuid",
        "output-variables" : ["some-var", "another-var"],
        "step": {
            "step_uuid": "step--a76dbc32-b739-427b-ae13-4ec703d5797e",
            "type": "action",
            "name": "IMC assets by CVE",
            "description": "Check the IMC for affected assets by CVE",
            "on_completion": "step--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4",
            "commands": [
                {
                    "type": "http-api",
                    "command": "GET http://__imc_address__/by/__cve__"
                }
            ]
        }
}
```

### Result

```plantuml
@startjson
{
       
        "execute-id" : "uuid",
        "executer-module-id" : "<your module id>",
        "status" : "ok|error|failed",
        "completion-time" : "2023-05-26T15:56:00.123456Z",
        "results" : [{ 
            "return-variable" : "<$$some-var$$>",
            "result-typ" : "bool|int|ip-address|uri|MACADDRESS|domain-name|custom",
            "result" : "<your result here>"
        },
        { 
            "return-variable" : "another-var",
            "result-typ" : "bool|int|ip-address|uri|mac-address|domain-name|custom",
            "result" : "<your result here>"
        }
        ]
}
```

#### Default schemas

* bool
* int
* ip-address
* uri
* mac-address
* domain-name


#### Example schema


## Sequences 

Example execution for SSH commands with SOARCA native capability. 


```plantuml
@startuml

participant Decomposer as decomposer
participant "Capability selector" as selector
participant "SSH executor" as ssh

decomposer -> selector : Execute(...)
alt capability in SOARCA
    selector -> ssh : execute ssh command
    ssh -> ssh : 
    selector <-- ssh : results
    decomposer <-- selector : OnCompletionCallback
else capability not available 
    decomposer <-- selector : Execution failure
    note right: No capability can handle command \nor capability crashed etc..
end
```

