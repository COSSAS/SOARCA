# SOARCA Database

SOARCA Database architecture

Mongo has different collections for SOARCA we use a database object per collection. So that would be:

* Workflow
* Coarse of Action
* Step


```plantuml
Interface IDatabase{
    void create(JsonData playbook)
    JsonData read(Id playbookId)
    void update(Id playbookId, JsonData playbook)
    void remove(Id playbookId)
}
Interface IWorkflowDatabase
Interface ICoaDatabase

class Controller
class WorkflowDatabase
class File
class Mongo


Controller -> IWorkflowDatabase
ICoaDatabase <- Controller
IWorkflowDatabase <|.. WorkflowDatabase
ICoaDatabase <|.. CoaDatabase
CoaDatabase -> IDatabase
IDatabase <- WorkflowDatabase
IDatabase <|.. File
IDatabase <|.. Mongo
```
## Getting data

### Getting workflow playbook data

```plantuml
participant Controller as controller
participant "Workflow Database" as workflow
database Database as db

controller -> workflow : get(id)
workflow -> db : read(playbookId)
note right
    When the create fails a error will be thrown
end note
workflow <-- db : "playbook JSON"
controller <-- workflow: "CacaoPlaybook Object"
```

### Writing workflow playbook data
```plantuml
participant Controller as controller
participant "Workflow Database" as workflow
database Database as db

controller -> workflow : set(CacaoPlaybook Object)
workflow -> db : create(playbook JSON)
note right
    When the create fails a error will be thrown
end note
```

### Update workflow playbook data
```plantuml
participant Controller as controller
participant "Workflow Database" as workflow
database Database as db

controller -> workflow : update(CacaoPlaybook Object)
workflow -> db : update(playbook id,playbook JSON)
note right
    When the create fails a error will be thrown
end note
workflow <-- db : true
controller <-- workflow: true
```

### Delete workflow playbook data
```plantuml

participant Controller as controller
participant "Workflow Database" as workflow
database Database as db

controller -> workflow : remove(playbook id)
workflow -> db : remove(playbook id)
note right
    When the create fails a error will be thrown
end note
```

### Handling an error
```plantuml
participant Controller as controller
participant "Workflow Database" as workflow
database Database as db

controller -> workflow : remove(playbook id)
workflow -> db : remove(playbook id)
workflow <-- db: error 
note right
    playbook does not exists
end note
controller <-- workflow: error 

```