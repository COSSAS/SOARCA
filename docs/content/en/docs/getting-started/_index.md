---
title: Getting Started
description: Getting SOARCA quickly setup
categories: [documentation, getting-started]
tags: [docker, bash]
weight: 2
date: 2024-03-18
---

## Prerequisites

Before you begin, you might need to install the following tools (Linux Ubuntu 22.04 adapt for your needs): 

- [golang](https://go.dev/doc/install)
- go gin `go get -u github.com/gin-gonic/gin`
- swaggo `go install github.com/swaggo/swag/cmd/swag@latest`
- cyclonedx-gomod `go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest`
- make `sudo apt install build-essential`
- [docker & docker compose](https://docs.docker.com/engine/install/)

## Quick Run

Below, we outline various options to kickstart SOARCA. The latest pre-compiled releases can be found [here](https://github.com/COSSAS/SOARCA/releases).

{{< tabpane langEqualsHeader=false  >}}
{{< tab header="make" lang="sh" >}}
make build && ./build/soarca
{{< /tab >}}
{{< tab header="Linux" lang="sh" >}}
wget https://github.com/COSSAS/SOARCA/releases/download/SOARCA_1.0.0/SOARCA_1.0.0_linux_amd64.tar.gz  && tar -xvf SOARCA* && ./SOARCA
{{< /tab >}}
{{< tab header="Docker Compose" lang="sh" >}}
cd docker/soarca && sudo docker compose up -d
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

Compiled binary files can be found under `/bin`. 


## Configuration

SOARCA reads its configuration from the environment variables or a `.env` file. An example of a `.env` is given below:

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

`docker pull cossas/soarca`

### Building from Source

```bash
git clone https://github.com/COSSAS/SOARCA.git
make build
cp .env.example .env
./build/soarca
```
