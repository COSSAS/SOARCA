# Decomposer

The decomposer will parse playbook objects to individual steps. This allows it to schedule new executor tasks. 

Each incoming playbook will be stored and executed individually. Decomposing is done to the step level. 


```plantuml

struct ExecutionDetails{
    uuid executionId 
    uuid playbookId
}

Interface IDecomposer{
ExecutionDetails, error Execute(cacao playbook)
error getStatus(uuid playbookId)    
}
Interface IExecuter

class Controller
class Decomposer

IDecomposer <- Controller
IDecomposer <|.. Decomposer
IExecuter -> Decomposer

```

```plantuml

Decomposer as decomposer

```

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
    note over exe: correct executer is selected
    ... Time has passed ...
    decomposer <-- exe
end loop

```