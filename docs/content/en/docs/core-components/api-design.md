---
title: API Description
description: >
    Descriptions for the SOARCA REST API endpoints 
categories: [API]
tags: [protocol, http, rest, api]
weight: 2
date: 2024-02-22
---

## Endpoint description

We will use HTTP status codes https://en.wikipedia.org/wiki/List_of_HTTP_status_codes


```plantuml
@startuml
protocol UiEndpoint {
    GET     /playbook
    GET     /playbook/meta
    POST    /playbook
    GET     /playbook/playbook-id
    PUT     /playbook/playbook-id
    DELETE  /playbook/playbook-id


    POST    /trigger/playbook
    POST    /trigger/playbook/id

    GET     /step

    GET     /status
    GET     /status/playbook
    GET     /status/playbook/id
    GET     /status/history

}
@enduml
```




### General messages

#### Error
When an error occurs a 400 status is returned with the following JSON payload, the original call can be omitted in production for security reasons.

responses: 400/Bad request

```plantuml
@startjson
{
    "status": "400",
    "message": "What went wrong.",
    "original-call": "<optional> Request JSON data",
    "downstream-call" : "<optional> downstream call JSON"
}
@endjson
```

#### Unauthorized
When the caller does not have valid authentication 401/unauthorized will be returned.


#### cacao playbook JSON

```plantuml
@startjson
{
            "type": "playbook",
            "spec_version": "cacao-2.0",
            "id": "playbook--91220064-3c6f-4b58-99e9-196e64f9bde7",
            "name": "coa flow",
            "description": "This playbook will trigger a specific coa",
            "playbook_types": ["notification"],
            "created_by": "identity--06d8f218-f4e9-4f9f-9108-501de03d419f",
            "created": "2020-03-04T15:56:00.123456Z",
            "modified": "2020-03-04T15:56:00.123456Z",
            "revoked": false,
            "valid_from": "2020-03-04T15:56:00.123456Z",
            "valid_until": "2020-07-31T23:59:59.999999Z",
            "derived_from": [],
            "priority": 1,
            "severity": 1,
            "impact": 1,
            "industry_sectors": ["information-communications-technology", "research", "non-profit"],
            "labels": ["soarca"],
            "external_references": [
                {
                    "name": "TNO SOARCA",
                    "description": "SOARCA Homepage",
                    "source": "TNO - COSSAS - HxxPS://LINK-TO-CODE-REPO.TLD",
                    "url": "HxxPS://LINK-TO-CODE-REPO.TLD",
                    "hash": "00000000000000000000000000000000000000000000000000000000000",
                    "external_id": "TNO/SOARCA 2023.01"
                }
            ],
            "features": {
                "if_logic": true,
                "data_markings": false
            },
            "markings": [],
            "playbook_variables": {
                "$$flow_data_location$$": {
                    "type": "string",
                    "value": "<mongodb_location>",
                    "description": "location of event and flow data",
                    "constant": true
                },
                "$$event_type$$": {
                    "type" : "string",
                    "value": "<event_type_string>",
                    "description": "type of incomming event / trigger",
                    "constant": true	
                }
            },
            "workflow_start": "step--d737c35f-595e-4abf-83ef-d0b6793556b9",
            "workflow_exception": "step--40131926-89e9-44df-a018-5f92f2df7914",
            "workflow": {
                "step--5ea28f63-ac32-4e5e-bd0c-757a50a3a0d7":{
                    "type": "single",
                    "name": "BI for CoAs",
                    "delay": 0,
                    "timeout": 30000,
                    "command": {
                        "type": "http-api",
                        "command": "hxxps://our.bi/key=VALUE"
                    },
                    "on_success": "step--71b15428-275a-49b5-9f09-3944972a0054",
                    "on_failure": "step--71b15428-275a-49b5-9f09-3944972a0054"
                },
                "step--71b15428-275a-49b5-9f09-3944972a0054": {
                    "type": "end",
                    "name": "End Playbook SOARCA Main Flow"
                }
            },
            "targets": { 

            },
            "extension_definitions": { }
        }
@endjson
```
---- 

### /playbook
The playbook endpoints are used to create playbooks in SOARCA, new playbooks can be added, and current ones edited and deleted. 

#### GET `/playbook`
Get all playbook ids that are currently stored in SOARCA.

##### Call payload
None

##### Response
200/OK with payload:

```plantuml
@startjson
[
    {
        "type": "playbook",
        "etc" : "etc..."   
    }
]
@endjson
```

##### Error
400/BAD REQUEST with payload:
General error

#### GET `/playbook/meta`
Get all playbook ids that are currently stored in SOARCA.

##### Call payload
None

##### Response
200/OK with payload:

```plantuml
@startjson

[
    {
        "id": "<playbook id>",
        "name": "<playbook name>",
        "description": "<playbook description>",
        "created": "<creation data time>",
        "valid_from": "<valid from date time>",
        "valid_until": "<valid until date time>",
        "labels": ["label 1","label 2"]
    }
]

@endjson
```

##### Error
400/BAD REQUEST with payload:
General error


#### POST `/playbook`
Create a new playbook and store it in SOARCA. The format is 


##### Payload

```plantuml
@startjson
{
    "type": "playbook",
    "etc" : "etc..."   
}
@endjson
```



##### Response
201/CREATED

```plantuml
@startjson
{
    "type": "playbook",
    "etc" : "etc..."   
}
@endjson
```

##### Error
400/BAD REQUEST with payload: General error, 409/CONFLICT if the entry already exists


#### GET `/playbook/{playbook-id}`
Get playbook details

##### Call payload
None

##### Response
200/OK with payload:

```plantuml
@startjson
{
    "<cacao-playbook> (json)"
}
@endjson
```
##### Error
400/BAD REQUEST

----

#### PUT `/playbook/{playbook-id}``
An existing playbook can be updated with PUT. 

##### Call payload
A playbook like [cacao playbook JSON](#cacao-playbook-json)


##### Response
200/OK with the edited playbook [cacao playbook JSON](#cacao-playbook-json)

##### Error
400/BAD REQUEST for malformed request

When updated it will return 200/OK or General error in case of an error.

----


#### DELETE `/playbook/{playbook-id}`
An existing playbook can be deleted with DELETE. When removed it will return 200/OK or general error in case of an error.

##### Call payload
None

##### Response
200/OK if deleted

##### Error
400/BAD REQUEST if the resource does not exist

---

#### POST `/trigger/playbook/xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx` 
Execute playbook with a specific id

##### Call payload
None

##### Response
Will return 200/OK when finished with playbook playbook.

```plantuml
@startjson
{
    "execution-id": "xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx",
    "playbook-id": "xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx"
}
@endjson
```

##### Error
400/BAD REQUEST general error on error.

---

#### POST `/trigger/playbook`
Execute an ad-hoc playbook

##### Call payload
A playbook like [cacao playbook JSON](#cacao-playbook-json)

##### Response
Will return 200/OK when finished with the playbook.

```plantuml
@startjson
{
    "execution-id": "xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx",
    "playbook-id": "xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx"
}
@endjson
```

##### Error
400/BAD REQUEST general error on error.

----

### /step [NOT in SOARCA V1.0]
Get capable steps for SOARCA to allow a coa builder to generate or build valid coa's

#### GET `/step`
Get all available steps for SOARCA. 

##### Call payload
None

##### Response
200/OK


```plantuml
@startjson
{
    
    "steps": [{
        "module": "executor-module",
        "category" : "analyses",
        "context" : "external",
        "step--5ea28f63-ac32-4e5e-bd0c-757a50a3a0d7":{
                    "type": "single",
                    "name": "BI for CoAs",
                    "delay": 0,
                    "timeout": 30000,
                    "command": {
                        "type": "http-api",
                        "command": "hxxps://our.bi/key=VALUE"
                    },
                    "on_success": "step--71b15428-275a-49b5-9f09-3944972a0054",
                    "on_failure": "step--71b15428-275a-49b5-9f09-3944972a0054"
                }}]
}
@endjson
```

Module is the executing module name that will do the executor call.

Category defines what kind of step is executed:
```plantuml
@startuml
enum workflowType {
    analyses
    action
    asset-look-up
    etc...
}
@enduml
```
Context will define whether the call is internal or external:

```plantuml
@startuml
enum workflowType {
    internal
    external
}
@enduml
```

##### Error
400/BAD REQUEST general error on error.

----

### /status
The status endpoints are used to get various statuses. 

#### GET `/status`
Call this endpoint to see if SOARCA is up and ready. This call has no payload body.

##### Call payload
None

##### Response
200/OK

```plantuml
@startjson
{
    "version": "1.0.0",
    "runtime": "docker/windows/linux/macos/other",
    "mode" : "development/production",
    "time" : "2020-03-04T15:56:00.123456Z",
    "uptime": {
        "since": "2020-03-04T15:56:00.123456Z",
        "milis": "uptime in miliseconds"
    }
}
@endjson
```

##### Error
5XX/Internal error, 500/503/504 message.

----

#### GET `/status/fins` | not implemented
Call this endpoint to see if SOARCA Fins are up and ready. This call has no payload body.

##### Call payload
None

##### Response
200/OK

```plantuml
@startjson
{
    "fins": [
        {
            "name": "Fin name",
            "status": "ready/running/failed/stopped/...",
            "id": "The fin UUID",
            "version": "semver verison: 1.0.0"
        }
    ]
}
@endjson
```

##### Error
5XX/Internal error, 500/503/504 message.


----

#### GET `/status/reporters` | not implemented
Call this endpoint to see which SOARCA reportes are used. This call has no payload body.

##### Call payload
None

##### Response
200/OK

```plantuml
@startjson
{
    "reporters": [
        {
            "name": "Reporter name"
        }
    ]
}
@endjson
```

##### Error
5XX/Internal error, 500/503/504 message.

----


#### GET `/status/ping`
See if SOARCA is up this will only return if all SOARCA services are ready

##### Call payload
None

##### Response
200/OK

`pong`

## Usage example flow

### Stand alone

```plantuml
@startuml
participant "SWAGGER" as gui
control "SOARCA API" as api
control "controller" as controller
control "Executor" as exe
control "SSH-module" as ssh


gui -> api : /trigger/playbook with playbook body
api -> controller : execute playbook playload
controller -> exe : execute playbook
exe -> ssh : get url from log
exe <-- ssh : return result
controller <-- exe : results
api <-- controller: results

@enduml
```

### Database load and execution

```plantuml
@startuml
participant "SWAGGER" as gui
control "SOARCA API" as api
control "controller" as controller
database "Mongo" as db
control "Executor" as exe
control "SSH-module" as ssh


gui -> api : /trigger/playbook/playbook--91220064-3c6f-4b58-99e9-196e64f9bde7
api -> controller : load playbook from database
controller -> db: retreive playbook
controller <-- db: playbook json
controller -> controller: validate playbook
controller -> exe : execute playbook
exe -> ssh : get url from log
exe <-- ssh : return result
controller <-- exe : results
api <-- controller: results

@enduml
```