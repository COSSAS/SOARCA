# SOARCA Executer Module

SOARCA is extendable by modules. Modules allow for new steps in playbook and added capability. 

## Requirements
Modules should be build in GO or Python and contain the following components.

. CACAO template to allow their capability to extend coarse of actions playbooks.
. MQTT protocol implementation to communicate with the SOARCA executor.
. Module specifies which `variables` it exposes for return types. These `variables` should be defined when submitting a module. 


## Native modules in SOARCA
The following capability modules are defined in SOARCA:
 
- SSH
- HTTP-API
- OPEN-C2

All modules have an well known GUID for there target definition. SOARCA will also extent the `agent-target-type-ov` with the following vocab for `ssh`, `http-api` and `openc2` respectively.

- soarca--00010001-1000-1000-a000-000100010001
- soarca--00020001-1000-1000-a000-000100010001
- soarca--00030001-1000-1000-a000-000100010001

The capability will be selected on the capability name and it must be unique.


### SSH capability
Well know guid: `soarca--00010001-1000-1000-a000-000100010001`

This module is defined in a playbook with the following TargetAgent definition:

```json
"agent_definitons": {
        "soarca--00010001-1000-1000-a000-000100010001": {
            "type": "soarca",
            "name": "soarca-ssh-capability"
        }
    },
```

This modules does not define variables as input. I will have the following output variables:

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
Well know guid: `soarca--00020001-1000-1000-a000-000100010001`

This module is defined in a playbook with the following TargetAgent definition:

```json
"agent_definitons": {
        "soarca--00020001-1000-1000-a000-000100010001": {
            "type": "soarca",
            "name": "soarca-http-api-capability"
        },
    },
```

```json
{
    "__soarca_http_result__": {
        Type: "string",
        Name: "result",
        Value: "<response from http-api here>"
    }
}
```

## OPEN-C2 capabilty
Well know guid: `soarca--00030001-1000-1000-a000-000100010001`

This module is defined in a playbook with the following TargetAgent definition:

```json
"agent_definitons": {
        "soarca--00030001-1000-1000-a000-000100010001": {
            "type": "soarca",
            "name": "soarca-open-c2-capability"
        },
    },
```

---

