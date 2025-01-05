---
title: Application design
weight: 1
categories: [architecture]
tags: [components]
description: >
    Details of the application architecture for SOARCA
---

##  Design decisions and core dependencies
To allow for fast execution and type-safe development SOARCA is developed in `go`. The application application can be deployed in `Docker`. Further dependencies are `MQTT` for the module system and `go-gin` for the REST API.


The overview on this page is aimed to guide you through the SOARCA architecture and components as well as the main flow. 

## Components

Components of SOARCA are displayed in the component diagram. 
- Green is implemented 
- Orange has limited functionality
- Red is not started but will be added in future releases

```plantuml
@startuml
set separator ::

protocol /playbook #lightgreen
protocol /trigger #lightgreen
protocol /status #lightgreen
protocol /reporter #lightgreen
protocol /manual #lightgreen
protocol /step #red
protocol /trusted/variables #red

class reporter::cache #orange
class core::decomposer #lightgreen
class core::executor  #lightgreen
class interaction::interaction  #lightgreen
class internal::controller  #lightgreen
class internal::database  #lightgreen
class internal::logger  #lightgreen
class api::playbook #lightgreen
class api::trigger #lightgreen
class core::capability::manual #lightgreen
class core::capability::http #lightgreen
class core::capability::ssh #lightgreen
class core::capability::openC2 #lightgreen
class core::capability::fin #lightgreen
class core::capability::powershell #lightgreen
class api::status #lightgreen
class api::reporter #lightgreen
class api::manual #lightgreen
class api::step #red
class api::variables #red



"/step" *-- api::step 
"/playbook" *-- api::playbook
"/trigger" *-- api::trigger
"/status" *-- api::status
"/reporter" *-- api::reporter
"/manual" *-- api::manual
"/trusted/variables" *-- api::variables

api *-down- internal::controller
internal::controller -* database
logger *- internal::controller
internal::controller -down-* core

api::reporter -down-> reporter::cache

reporter::cache <-- core::decomposer
reporter::cache <-- core::executor

api::playbook --> internal::database
api::trigger --> core::decomposer
api::trigger --> internal::database
api::manual --> interaction::interaction

core::decomposer -down-> core::executor
core::executor --> core::capability::openC2
core::executor --> core::capability::fin
core::executor --> core::capability::http
core::executor --> core::capability::ssh
core::executor --> core::capability::powershell
core::executor --> core::capability::manual
interaction::interaction <-- core::capability::manual
@enduml
```

## Classes

This diagram consists of the class structure used by SOARCA

```plantuml
@startuml

interface IPlaybook
interface IStatus
interface ITrigger
Interface IPlaybookDatabase
Interface IDatabase
Interface ICapability
Interface IDecomposer
Interface IExecutor

class Controller
class Decomposer
class PlaybookDatabase
class Status
class Mongo
class Capability
Class Executor


IPlaybook <|.. Playbook
ITrigger <|.. Trigger
IStatus <|.. Status
ICapability <|.. Capability
IExecutor <|.. Executor
Trigger -> IPlaybookDatabase
IPlaybookDatabase <- Playbook
IPlaybookDatabase <|.. PlaybookDatabase
IDatabase <-up- PlaybookDatabase
IDatabase <|.. Mongo
IDecomposer <- Trigger
IDecomposer <|.. Decomposer
IExecutor <- Decomposer
ICapability <- Executor
@enduml
```

### Controller
The SOARCA controller will create all classes needed by SOARCA. The controller glues the api and decomposer together. Each run will instantiate a new decomposer. 

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
Interface IExecutor

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
IExecutor -> Decomposer

```



## Main application flow
These sequences will show a simplified overview of how the SOARCA components interact.

The main flow of the application is the following. Execution will start by processing the JSON formatted CACAO playbook if successful the playbook is handed over to the Decomposer. This is where the playbook is decomposed into its parts and passed step by step to the executor. These operations will block the API until execution is finished. For now, no variables are exposed via the API to the caller.

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