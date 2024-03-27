---
title: Reporting
weight: 8
categories: [architecture]
tags: [apis]
description: >
    Reporting of Playbook worfklow information and steps execution
---

SOARCA utilizes push-based reporting to provide information on the instantiation of a CACAO workflow, and information on the execution of workflow steps.


## General Reporting Architecture

For the execution of a playbook, a *Decomposer* and invoked *Executor*s are injected with a *Reporter*. The *Reporter* maintains the reporting logic that reports execution information to a set of specified and available targets.

A reporting target can be internal to SOARCA, such as the [database](database.md), and the [report] API. A reporting target can also be a third party tool, such as an external SOAR, or incident case management system.

Upon execution trigger for a playbook, information about the chain of playbook steps to be executed will be pushed to the targets via dedicated reporting classes.

Along the execution of the workflow steps, the reporting classes will dynamically update the steps execution information such as output variables, and step execution success or failure.

The reporting capabilities will thus allow to populate and update views and information about a workflow composition, and its dynamic execution outcomes, to SOARCA internal as well as third party tools.

The schema below represents the architecture concept.


```plantuml
@startuml
set separator ::

interface IDecomposerReporter{
    ReportWorkflow(cacao.workflow)
}

interface IExecuterReporter{
    ReportStep(cacao.workflow.Step)
}

class Reporter 
class DB
class 3PTool
class Decomposer
class Executor

Decomposer -up-> Reporter
Executor -up-> Reporter

Reporter .up.|> IDecomposerReporter
Reporter -up-> IDecomposerReporter
Reporter .up.|> IExecuterReporter
Reporter -up-> IExecuterReporter

DB .up.|> IDecomposerReporter
DB .up.|> IExecuterReporter
3PTool .up.|> IDecomposerReporter
3PTool .up.|> IExecuterReporter

Reporter -left-> DB
Reporter -right-> 3PTool

```

### Interfaces

The logic and extensibility is implemented in the SOARCA architecture by means of reporting interfaces. At this stage, we implement an *IDecomposerReporter* to push information about the  entire workflow to be executed, and an *IExecuterReporter* to push step-specific information as the steps of the workflow are executed.

A high level *Reporter* class will implement both interfaces, and maintain the list of decomposer and executor reporters activated  for the SOARCA instance. The *Reporter* class will invoke all reporting functions for each active reporter.

## Future plans

At this stage, third party tools integrations may be built in SOARCA via packages implementing reporting logic for the specific tools. Alternatively, third party tools may implement pull-based mechanisms to get information from the execution of a playbook via SOARCA.

In the near future, we will (also) make available a SOARCA Report API that can establish a WebSocket connection to a third party tool. This will thus allow SOARCA to push execution updates as they come to third party tools, without external tools having to poll SOARCA.