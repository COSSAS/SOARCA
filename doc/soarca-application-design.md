# Application design

The application consist of the endpoint which control the workflows/ Coarse of Actions and steps that are available.

## Component 

```plantuml
@startuml
set separator ::

protocol /stix
protocol /coa
protocol /step
protocol /workflow
protocol /trusted/variables
protocol /status

class endpoints::api
class endpoints::taxiiconsumer
class controller
class database
class log

"/coa" *-- endpoints::coa
"/step" *-- endpoints::step
"/workflow" *-- endpoints::workflow
"/status" *-- endpoints::api
"/trusted/variables" *-- endpoints::api
"/stix" *-- endpoints::taxiiconsumer

endpoints -down-> controller
controller -* database
log *- controller
controller -down-* cacao::executer
class cacao::modules::openC2
class cacao::modules::http
cacao::executer --> cacao::modules::openC2
cacao::executer --> cacao::modules::http
cacao::executer --> cacao::modules::asset
cacao::executer --> cacao::modules::ssh
@enduml
```

## Classes

This diagram consists of the class structure used by SOAR-CA

```plantuml
@startuml

interface IWorkflow
interface ICoa
interface IStatus
interface ITrigger
Interface IWorkflowDatabase
Interface IDatabase
Interface IQueue
Interface IDecomposer
Interface IExecuter

class Controller
class Decomposer
class WorkflowDatabase
class File
class Mongo
class Queue
Class Executer


IWorkflow <|.. Controller
ICoa <|.. Controller
ITrigger <|.. Controller
IStatus <|.. Controller
IQueue <|.. Queue
IExecuter <|.. Executer
Controller -> IWorkflowDatabase
IWorkflowDatabase <|.. WorkflowDatabase
IDatabase <-up- WorkflowDatabase
IDatabase <|.. File
IDatabase <|.. Mongo
IDecomposer <- Controller
IDecomposer <|.. Decomposer
IExecuter -> Decomposer
IQueue -> Executer
@enduml
```
