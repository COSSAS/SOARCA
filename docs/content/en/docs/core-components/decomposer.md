---
title: Decomposer
weight: 4
categories: [architecture]
tags: []
description: >
  Playbook deconstructor architecture
---


## Decomposer structure
The decomposer will parse playbook objects to individual steps. This allows it to schedule new executor tasks. 

Each incoming playbook will executed individually. Decomposing is done up to the step level.

{{% alert title="Warning" color="warning" %}}
SOARCA 1.0.x will only support steps of type `action`
{{% /alert %}}

```plantuml

struct ExecutionDetails{
    uuid executionId 
    uuid playbookId
}

Interface IDecomposer{
ExecutionDetails, error Execute(cacao playbook)
error getStatus(uuid playbookId)    
}
Interface IExecutor

class Controller
class Decomposer

IDecomposer <- Controller
IDecomposer <|.. Decomposer
IExecutor <- Decomposer

```

### IExecutor
Interface for interfacing with the Executor this will in turn select and execute the command on the right [module](/docs/core-components/modules) or [fin](/docs/soarca-extensions/).

### Execution details
The struct contains the details of the execution (execution id which is created for every execution) and the playbook id. The combination of these is unique. 

## Decomposition of playbook


```plantuml
participant caller 
participant "Playbook decomposer" as decomposer
participant "Playbook state" as queue
participant Executor as exe



caller -> decomposer: Execute
caller <-- decomposer: ExecutionStatus

decomposer -> queue: store state
decomposer <-- queue:

loop for all playbook steps
    decomposer -> queue: load state
    decomposer <-- queue: 
    decomposer -> decomposer: parse step
    decomposer -> decomposer: parse command
   
    decomposer -> exe: execute command
    note over exe: correct executor is selected
    ... Time has passed ...
    decomposer <-- exe
end loop

```