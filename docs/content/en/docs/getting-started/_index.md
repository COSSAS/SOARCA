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
cd docker/soarca && docker compose up -d
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

curl -X POST -H "Content-Type: application/json" -d @./example/ssh-playbook.json localhost:8080/trigger/playbook
{{< /tab >}}
{{< tab header="http" lang="sh" >}}
curl -X POST -H "Content-Type: application/json" -d @./example/http-playbook.json localhost:8080/trigger/playbook
{{< /tab >}}
{{< tab header="openC2" lang="sh" >}}
curl -X POST -H "Content-Type: application/json" -d @./example/openc2-playbook.json localhost:8080/trigger/playbook
{{< /tab >}}
{{< /tabpane >}}

### Caldera setup

SOARCA optionally comes packaged together with Caldera. To use the
[Caldera capability](/docs/soarca-extensions/native-capabilities#caldera-capability), simply make
sure you use the right Compose file when running:

```diff
- cd docker/soarca && docker compose up -d
+ cd docker/soarca && docker compose --profile caldera up -d
```

{{% alert title="Warning" %}}
This only works when using Docker Compose to run SOARCA. When building SOARCA from scratch,
you should supply your own Caldera instance and [configure](#configuration) its URL manually.
{{% /alert %}}

## Configuration

SOARCA reads its configuration from the environment variables or a `.env` file. An example of a `.env` is given below:

{{< tabpane langEqualsHeader=false  >}}
{{< tab header="`.env`" lang="txt" >}}
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

CALDERA_URL: ""

HTTP_SKIP_CERT_VALIDATION: false
{{< /tab >}}
{{< /tabpane >}}

The environment variables have the following meaning:

|variable |content |description
|---|---|---|
|PORT |port  |Set the exposed port of SOARCA the default is `8080`
|DATABASE |true \| false   | Set if you want to run with external database default is `false`
|MONGODB_URI |uri  |Set the Mongo DB uri default is `mongodb://localhost:27017`
|DATABASE_NAME |name  |Set the Mongo DB database name when using docker default is `soarca`
|DB_USERNAME |user  |Set the Mongo DB database user when using docker default is `root`
|DB_PASSWORD |password  |Set the Mongo DB database users password when using docker default is `rootpassword`. IT IS RECOMMENDED TO CHANGE THIS IN PRODUCTION!
|MAX_REPORTERS |number  |Set the maximum number of downstream reporters default is `5` 
|LOG_GLOBAL_LEVEL |[Log levels]  |One of the specified log levels. Defaults to `info`
|LOG_MODE |development \| production  |If production is chosen the `LOG_GLOBAL_LEVEL` is used for all modules defaults to `production`
|LOG_FILE_PATH |filepath  |Path to the logfile you want to use for all logging. Defaults to `""` (empty string)
|LOG_FORMAT |text \| json  |The logging can be in plain text format or in JSON format. Defaults to `json`
|MQTT_BROKER | dns name or ip | The broker address for SOARCA to connect to, for communication with fins default is `localhost`
|MQTT_PORT   | port | The broker address for SOARCA to connect to, for communication with fins default is `1883`
|ENABLE_FINS| true \| false | Enable fins in SOARCA defaults to `false`
|CALDERA_URL| url | Instance URL which the [Caldera capability](/docs/soarca-extensions/native-capabilities#caldera-capability) may use; leaving this empty will disable the Caldera capability
|VALIDATION_SCHEMA_URL|url| Set a custom validation schema to be used to validate playbooks defaul is `""` to use internal. NOTE: changing this heavily impacts performance. 

## Obtaining

There are several ways to obtain a copy of the SOARCA software.

### Docker Hub 

A prebuilt image can be pulled from the
[Docker Hub](https://hub.docker.com/r/cossas/soarca):

```bash
docker pull cossas/soarca
```

### Building from source

```bash
git clone https://github.com/COSSAS/SOARCA.git
make build
cp .env.example .env
./build/soarca
```
