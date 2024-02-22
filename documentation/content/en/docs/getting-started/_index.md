---
title: Getting Started
description: Getting SOARCA quickly setup
categories: [documentation, getting-started]
tags: [docker, bash, ]
weight: 2
date: 2023-01-05
---

## Prerequisites

Before you begin, you might need to install the following tools:


- [golang](https://go.dev/doc/install)
- [docker & docker compose](https://docs.docker.com/engine/install/)

## Quick Run

Below, we outline various options to kickstart SOARCA. The latests pre-compiled releases can be found [here]().

{{< tabpane langEqualsHeader=false  >}}
{{< tab header="make" lang="sh" >}}
make run
{{< /tab >}}
{{< tab header="Linux" lang="sh" >}}
wget .... && chmod +x {}
{{< /tab >}}
{{< tab header="Docker" lang="sh" >}}
sudo docker compose up -d
{{< /tab >}}
{{< tab header="Docker Compose" lang="sh" >}}
make run
{{< /tab >}}
{{< /tabpane >}}




{{% alert title="Tip" %}}
Output will be similar to:
{{< tabpane langEqualsHeader=false  >}}
{{< tab header="make" lang="sh" >}}
swag init
2024/02/09 12:53:04 Generate swagger docs....
2024/02/09 12:53:04 Generate general API Info, search dir:./
2024/02/09 12:53:06 Generating cacao.Playbook
2024/02/09 12:53:06 Generating cacao.ExternalReferences
2024/02/09 12:53:06 Generating cacao.Workflow
2024/02/09 12:53:06 Generating cacao.Step
.....
{{< /tab >}}
{{< /tabpane >}}
{{% /alert %}}

Compiles binary files can be found under `/bin`. 


## Configuration

SOARCA reads it's configuration from the environment variables or an `.env` file. An example of an `.env` is given below:

{{< tabpane langEqualsHeader=false  >}}
{{< tab header="`.env`" lang="txt" >}}
PORT: 8080
MONGODB_URI: "mongodb://localhost:27017"
DATABASE_NAME: "soarca"
DB_USERNAME: "root"
DB_PASSWORD: "rootpassword"
WORKFLOW_API_LOG_LEVEL: trace
DATABASE: "false"

LOG_GLOBAL_LEVEL: "info"
LOG_MODE: "development"
LOG_FILE_PATH: ""
LOG_FORMAT: "json"
{{< /tab >}}
{{< /tabpane >}}

### Docker hub 

`<to be added soon>`

### Building from Source

`<to be added soon>`
