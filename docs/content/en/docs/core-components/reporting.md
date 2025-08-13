---
title: Reporter
weight: 8
categories: [architecture]
tags: [api]
description: >
    Reporting of Playbook worfklow information and steps execution
---

SOARCA utilizes push-based reporting to provide information on the instantiation of a CACAO workflow, and information on the execution of workflow steps.


## General Reporting Architecture

For the execution of a playbook, a *Decomposer* and invoked *Executor*s are injected with a *Reporter*. The *Reporter* maintains the reporting logic that reports execution information to a set of specified and available targets.

A reporting target can be internal to SOARCA, such as a [Cache](#cache-reporter). A reporting target can also be a third-party tool, such as an external SOAR/ SIEM, or incident case management system.

Upon execution trigger for a playbook, information about the chain of playbook steps to be executed will be pushed to the targets via dedicated reporting classes.

Along the execution of the workflow steps, the reporting classes will dynamically update the steps execution information such as output variables, and step execution success or failure.

The reporting features will enable the population and updating of views and data concerning workflow composition and its dynamic execution results. This data can be transmitted to SOARCA internal reporting components such as a cache, as well as to third-party tools.

The schema below represents the architecture concept.


```plantuml
@startuml
set separator ::

interface IStepReporter{
   ReportStep() error
}

interface IWorkflowReporter{
    ReportWorkflow() error
}


interface IDownStreamReporter {
    ReportWorkflow() error
	ReportStep() error
}

class Reporter {
    reporters []IDownStreamReporter

    RegisterReporters() error
    ReportWorkflow()
    ReportStep()
}

class Cache
class 3PTool

class Decomposer
class Executor

Decomposer -right-> IWorkflowReporter
Executor -left-> IStepReporter

Reporter .up.|> IStepReporter
Reporter .up.|> IWorkflowReporter
Reporter -right-> IDownStreamReporter

Cache .up.|> IDownStreamReporter
3PTool .up.|> IDownStreamReporter

```

### Interfaces

The reporting logic and extensibility is implemented in the SOARCA architecture by means of reporting interfaces. At this stage, we implement an *IWorkflowReporter* to push information about the entire workflow to be executed, and an *IStepReporter* to push step-specific information as the steps of the workflow are executed.

A high level *Reporter* component will implement both interfaces, and maintain the list of *DownStreamRepporter*s activated for the SOARCA instance. The *Reporter* class will invoke all reporting functions for each active reporter. The *Executor* and *Decomposer* components will be injected each with the Reporter though, as interface of respectively workflow reporter, and step reporter, to keep the reporting scope separated.

The *DownStream* reporters will implement push-based reporting functions specific for the reporting target, as shown in the *IDownStreamReporter* interface. Internal components to SOARCA, and third-party tool reporters, will thus implement the *IDownStreamReporter* interface.

## Future plans

At this stage, third-party tools integrations may be built in SOARCA via packages implementing reporting logic for the specific tools. Alternatively, third-party tools may implement pull-based mechanisms (via the API) to get information from the execution of a playbook via SOARCA.

In the near future, we will (also) make available a SOARCA Report API that can establish a WebSocket connection to a third-party tool. As such, this will thus allow SOARCA to push execution updates as they come to third-party tools, without external tools having to poll SOARCA.

## Native Reporters

SOARCA implements internally reporting modules to handle database and caches reporting.

### Cache reporter

The *Cache* reporter mediates between [decomposer](https://cossas.github.io/SOARCA/docs/core-components/decomposer/) and [executors](https://cossas.github.io/SOARCA/docs/core-components/executor/), [database](https://cossas.github.io/SOARCA/docs/core-components/database/), and reporting APIs. As *DownStreamReporter*, the *Cache* stores workflow and step reports in-memory for an ongoing execution. As *IExecutionInformant*, the *Cache* provides information to the reporting API. The schema below shows how it is positioned in the SOARCA architecture.

```plantuml
@startuml

protocol /reporter

interface IDownStreamReporter {
    ReportWorkflow() error
	ReportStep() error
}
interface IDatabase

interface IExecutionInformer

class ReporterApi
class Reporter
class Cache {
    cache []ExecutionEntry
}


"/reporter" -right-> ReporterApi

Reporter -> IDownStreamReporter
Cache -left-> IDatabase

Cache .up.|> IDownStreamReporter
Cache .up.|> IExecutionInformer
ReporterApi -down-> IExecutionInformer

```

The *Cache* thus reports the execution information downstream both in the database, and in memory. Upon execution information requests from the `/reporter` API, the cache can provide information fetching either from memory, or querying the database.