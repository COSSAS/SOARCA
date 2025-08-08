
---
title:  Playbook conversion
description: >
    Convert playbooks in other formats to CACAO
categories: []
tags: []
weight: 2
date: 2025-08-08
---

## Design
We offer basic support to convert playbooks in other formats to CACAO.
The only currently supported format is BPMN.

## Usage
The conversion tool is implemented in Go as a standalone executable.
The executable accepts three flag arguments: 
 - "source" for the file that is to be converted
 - "target" for the name of the converted CACAO playbook file (optional, defaults to appending ".json" to the source file)
 - "format" for the format of the source file (optional, but if the format is not clear this is required)

{{< tabpane langEqualsHeader=false  >}}
{{< tab header="make" lang="sh" >}}
make build 
./build/soarca-conversion -source=... [-target=...] [-format=...]
{{< /tab >}}
{{< /tabpane >}}

## BPMN
For examples of BPMN playbooks look [here](https://github.com/cisagov/shareable-soar-workflows/tree/develop).
BPMN playbooks lack the explicit actions that are possible within CACAO playbooks, and converted files are thus not immediately useful.

Certain BPMN features are not possible to implement in CACAO. These are:
 - intermediateCatchEvent, intermediateThrowEvent: these can be implemented by using the playbook step in CACAO, but this will need to be done by the implementer.
 - inclusiveGateway: these can be implemented using switch nodes in CACAO, which are currently not implemented.

To successfully convert a BPMN playbook to a functional CACAO playbook, take the following steps:
 - Apply the conversion tool
 - Update playbook metadata and agent definitions if applicable
 - Update actions to perform any actions that are possible to express in CACAO (ssh, http, etc.)
 - Update the playbook to use variables to make if-else constructions work as intended.
 - Convert any constructions that are not in BPMN but that are possible in CACAO, such as while-loops.
