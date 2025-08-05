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
{{< tab header="Docker Compose" lang="sh" >}}
cd deployments/docker/soarca && sudo docker compose up -d
{{< /tab >}}
{{< tab header="make" lang="sh" >}}
make build && ./build/soarca
{{< /tab >}}
{{< tab header="Linux" lang="sh" >}}
wget https://github.com/COSSAS/SOARCA/releases/download/SOARCA_1.0.0/SOARCA_1.0.0_linux_amd64.tar.gz  && tar -xvf SOARCA* && ./SOARCA
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

### Playbook execution

You can use the following commands to execute the example playbooks via the terminal while SOARCA is running assuming on localhost. Alternatively you can go to `http://localhost:8080/swagger/index.html` and use the trigger/playbook endpoint.

Example playbooks:
{{< tabpane langEqualsHeader=false  >}}
{{< tab header="ssh" lang="sh" >}}

# make sure an ssh server is running on adres 192.168.0.10

curl -X POST -H "Content-Type: application/json" -d @./examples/ssh-playbook.json localhost:8080/trigger/playbook
{{< /tab >}}
{{< tab header="http" lang="sh" >}}
curl -X POST -H "Content-Type: application/json" -d @./examples/http-playbook.json localhost:8080/trigger/playbook
{{< /tab >}}
{{< tab header="openC2" lang="sh" >}}
curl -X POST -H "Content-Type: application/json" -d @./examples/openc2-playbook.json localhost:8080/trigger/playbook
{{< /tab >}}
{{< /tabpane >}}

## Configuration

SOARCA reads its configuration from the environment variables or a `.env` file. An example of a `.env` is given below:

{{< tabpane langEqualsHeader=false  >}}
{{< tab header="`.env`" lang="txt" >}}
HOST: localhost
PORT: 8080
SOARCA_ALLOWED_ORIGINS: "*"
GIN_MODE: "release"
MONGODB_URI: "mongodb://localhost:27017"
DATABASE_NAME: "soarca"
DB_USERNAME: "root"
DB_PASSWORD: "rootpassword"
PLAYBOOK_API_LOG_LEVEL: trace
DATABASE: "false"
MAX_REPORTERS: "5"

LOG_GLOBAL_LEVEL: "info"
LOG_MODE: "development"
LOG_FILE_PATH: ""
LOG_FORMAT: "json"

ENABLE_FINS: false
MQTT_BROKER: "localhost"
MQTT_PORT: 1883

HTTP_SKIP_CERT_VALIDATION: false
{{< /tab >}}
{{< /tabpane >}}


For more custom and advanced deployment instructions go [here](/docs/installation-configuration/_index.md).
### Docker hub

`docker pull cossas/soarca`

### Building from Source

```bash
git clone https://github.com/COSSAS/SOARCA.git
make build
cp .env.example .env
./build/soarca
```
