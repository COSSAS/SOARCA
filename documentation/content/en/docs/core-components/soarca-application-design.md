---
title: SOARCA Application design
weight: 1
categories: [architecture]
tags: [components]
description: >
    The application consist of the endpoint which control the playbooks/ Coarse of Actions and steps that are available.
---

##  Design decisions and core dependencies
To allow for fast execution and type safe development SOARCA is developed in `go`. The application application can be deployed in `Docker`. Further dependencies are `MQTT` for the module system and `go-gin` for the REST API.


The overview on this page is aimed to guid you through the SOARCA architecture and components as well as the main flow. 

## Components

Components of SOARCA are displayed in the component diagram. 
- Green is implemented, 
- orange has limited functionality, 
- red is not started but will be added in future releases.

```plantuml
@startuml
set separator ::

protocol /playbook #lightgreen
protocol /trigger #lightgreen
protocol /step #red
protocol /trusted/variables #red
protocol /status #red

class controller  #lightgreen
class database  #lightgreen
class log  #lightgreen
class core::decomposer #lightgreen
class core::executor  #lightgreen
class endpoints::playbook #lightgreen
class endpoints::trigger #lightgreen
class core::modules::http #lightgreen
class core::modules::ssh #lightgreen
class core::modules::openC2 #orange
class core::modules::fin #orange

class endpoints::step #red
class endpoints::variables #red
class endpoints::status #red


"/step" *-- endpoints::step 
"/playbook" *-- endpoints::playbook
"/trigger" *-- endpoints::trigger
"/status" *-- endpoints::status
"/trusted/variables" *-- endpoints::variables

endpoints *-down- controller
controller -* database
log *- controller
controller -down-* core::decomposer
core::decomposer -down-> core::executor
core::executor --> core::modules::openC2
core::executor --> core::modules::fin
core::executor --> core::modules::http
core::executor --> core::modules::ssh
@enduml
```

## Classes

This diagram consists of the class structure used by SOAR-CA

```plantuml
@startuml

interface IPlaybook
interface IStatus
interface ITrigger
Interface IPlaybookDatabase
Interface IDatabase
Interface ICapability
Interface IDecomposer
Interface IExecuter

class Controller
class Decomposer
class PlaybookDatabase
class Status
class Mongo
class Capability
Class Executer


IPlaybook <|.. Playbook
ITrigger <|.. Trigger
IStatus <|.. Status
ICapability <|.. Capability
IExecuter <|.. Executer
Trigger -> IPlaybookDatabase
IPlaybookDatabase <- Playbook
IPlaybookDatabase <|.. PlaybookDatabase
IDatabase <-up- PlaybookDatabase
IDatabase <|.. Mongo
IDecomposer <- Trigger
IDecomposer <|.. Decomposer
IExecuter <- Decomposer
ICapability <- Executer
@enduml
```

### Controller
The SOARCA controller will create all classed needed by SOARCA. The controller glues the endpoints and decomposer together. Each run will instantiate a new decomposer. 

```plantuml
interface IPlaybook{
    void Get()
    void Get(PlaybookId id)
    void Add(Playbook playbook)
    void Update(Playbook playbook)
    void Remove(Playbook playbook)
}

interface IStatus{

}

interface ITrigger{
    void TriggerById(PlaybookId id)
    void Trigger(Playbook playbook)
}

Interface IPlaybookDatabase

Interface IDecomposer
Interface IExecuter

class Trigger
class Controller
class Decomposer

IPlaybook <|.. Playbook
ITrigger <|.. Trigger
IStatus <|.. Status


Trigger -> IPlaybookDatabase
IPlaybookDatabase <- Playbook
IPlaybookDatabase <|.. PlaybookDatabase


IDecomposer <- Trigger
IDecomposer <|.. Decomposer
IExecuter -> Decomposer

```



## Main application flow
These sequences will show the simplified overview how the SOARCA components interact.

The main flow of the application is the following. Execution will start by processing the JSON formatted CACAO playbook if successful the playbook is handed over to the Decomposer. This is where the playbook is decomposed into it's parts and passed step by step to the executor. These operations will block the api until execution is finished. For now no variables are exposed via the API to the caller.

```plantuml
Actor Caller
Caller -> Api
Api -> Trigger : /trigger
Trigger -> Decomposer : Trigger playbook as ad-hoc execution
loop for each step 
Decomposer -> Executor : Send step to executor
Executor -> Executor : select capability (ssh selected)
Executor -> Ssh : Command
Executor <-- Ssh : return
Decomposer <-- Executor
else execution failure (break loop)
Executor <-- Ssh : error
Decomposer <-- Executor: error
Decomposer -> Decomposer : stop execution

end 
Trigger <-- Decomposer : execution details
Api <-- Trigger : execution details
Caller <-- Api
```