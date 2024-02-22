---
title: SOARCA Database
weight: 7
categories: [architecture]
tags: [database]
description: >
  Database details of SOARCA
---

SOARCA Database architecture, SOARCA makes use of [MongoDB](https://www.mongodb.com). It is used to store and retrieve playbooks. Later it will also store individual steps.

## Mongo

Mongo has different collections for SOARCA we use a database object per collection. So that would be:

* Playbook
* Step


```plantuml
Interface IDatabase{
    void create(JsonData playbook)
    JsonData read(Id playbookId)
    void update(Id playbookId, JsonData playbook)
    void remove(Id playbookId)
}
Interface IPlaybookDatabase
Interface IStepDatabase

class Controller
class PlaybookDatabase

class Mongo


Controller -> IPlaybookDatabase
IStepDatabase <- Controller
IPlaybookDatabase <|.. PlaybookDatabase
IStepDatabase <|.. StepDatabase
StepDatabase -> IDatabase
IDatabase <- PlaybookDatabase

IDatabase <|.. Mongo
```
## Getting data

### Getting playbook data

```plantuml
participant Controller as controller
participant "Playbook Database" as playbook
database Database as db

controller -> playbook : get(id)
playbook -> db : read(playbookId)
note right
    When the create fails a error will be thrown
end note
playbook <-- db : "playbook JSON"
controller <-- playbook: "CacaoPlaybook Object"
```

### Writing playbook data
```plantuml
participant Controller as controller
participant "Playbook Database" as playbook
database Database as db

controller -> playbook : set(CacaoPlaybook Object)
playbook -> db : create(playbook JSON)
note right
    When the create fails a error will be thrown
end note
```

### Update playbook data
```plantuml
participant Controller as controller
participant "Playbook Database" as playbook
database Database as db

controller -> playbook : update(CacaoPlaybook Object)
playbook -> db : update(playbook id,playbook JSON)
note right
    When the create fails a error will be thrown
end note
playbook <-- db : true
controller <-- playbook: true
```

### Delete playbook data
```plantuml

participant Controller as controller
participant "Playbook Database" as playbook
database Database as db

controller -> playbook : remove(playbook id)
playbook -> db : remove(playbook id)
note right
    When the create fails a error will be thrown
end note
```

### Handling an error
```plantuml
participant Controller as controller
participant "Playbook Database" as playbook
database Database as db

controller -> playbook : remove(playbook id)
playbook -> db : remove(playbook id)
playbook <-- db: error 
note right
    playbook does not exists
end note
controller <-- playbook: error 

```