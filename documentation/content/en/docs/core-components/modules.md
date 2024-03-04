---
title: Executer Modules
weight: 6
categories: [architecture]
tags: [components]
description: >
    Native executer modules 
---

## Requirements
Executer modules are part of the SOARCA core. Executer modules perform the actual commands in CACAO playbook steps.


## Native modules in SOARCA
The following capability modules are defined in SOARCA:
 
- ssh
- http-api
- openc2-http

The capability will be selected on the type of the agent in the CACAO playbook step. This type must be equal to `soarca-<capability identifier>`.

### SSH capability

This module is defined in a playbook with the following TargetAgent definition:

```json
"agent_definitons": {
        "soarca--00010001-1000-1000-a000-000100010001": {
            "type": "soarca-ssh"
        }
    },
```

This modules does not define specific variables as input, but of course variable interpolation is supported in the command and target definitions. It has the following output variables:

```json
{
    "__soarca_ssh_result__": {
        Type: "string",
        Name: "result",
        Value: "<output from command here>"
    }
}
```

If the connection to the target fail the structure will be set but be empty and an error will be returned. If no error occurred nil is returned.


## HTTP-API capability

This module is defined in a playbook with the following TargetAgent definition:

```json
"agent_definitons": {
        "soarca--00020001-1000-1000-a000-000100010001": {
            "type": "soarca-http-api"
        },
    },
```

It supports variable interpolation in the command, port, authentication info, and target definitions.

The result of the step is stored in the following output variables:

```json
{
    "__soarca_http_api_result__": {
        Type: "string",
        Name: "result",
        Value: "<response from http-api here>"
    }
}
```

## OPEN-C2 capabilty

This module is defined in a playbook with the following TargetAgent definition:

```json
"agent_definitons": {
        "soarca--00030001-1000-1000-a000-000100010001": {
            "type": "soarca-openc2-http"
        },
    },
```

It supports variable interpolation in the command, headers, and target definitions.

The result of the step is stored in the following output variables:

```json
{
    "__soarca_openc2_http_result__": {
        Type: "string",
        Name: "result",
        Value: "<response from openc2-http here>"
    }
}
```

---

## MQTT fin module
This module is used by SOARCA to communicate with fins (capabilities) see [fin documentation](/docs/soarca-extentions/) for more information
