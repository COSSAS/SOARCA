---
title: Executable playbooks
weight: 3
description: >
  A playbook primer
resources:
- src: "*Slide.png"
  params:
    byline: "*Slide*: © 2024 TNO"
---

SOARCA is build on top of the [CACAO Security Playbook Version 2.0](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/cs01/security-playbooks-v2.0-cs01.html) standard.

{{% alert title="Warning" color="warning" %}}
SOARCA v1.0 only implements a part of the CACAO v2 spec. Only `start`, `end`, and `action` steps are supported at this time.
{{% /alert %}}

A CACAO playbook is a structured document that outlines a series of orchestrated actions to address specific security events, incidents, or other security-related activities. These playbooks allow for the automation of security steps.

SOARCA is a _security orchestrator_ that reads the steps defined in a CACAO playbook and performs the necessary actions to execute the commands they contain. This makes a CACAO document an _executable playbook_.

SOARCA's development is ongoing, and at this time, it only partly supports the entire CACAO specification. On this page we'll go over the general concepts in a CACAO playbooks and the parts of the standard that are supported by SOARCA.

## A CACAO playbook

Here, we have an example of a relatively simple CACAO playbook that demonstrates [SOARCA's capabilities](/docs/soarca-extensions). The flow of steps is depicted in the following image:

![Example playbook flow](/SOARCA/images/example-playbook.svg)

The JSON of this CACAO playbook can be found [at the bottom of this page](#example-playbook).

As you can see, this playbook contains a mix of logical steps (`if-condition` and `while-condition`) and `action` steps that perform commands on a target system.

### Agents and targets

In CACAO playbooks, entities that execute commands are called agents, and the entities against which the commands are executed are called targets (see [Agents and Targets](https://docs.oasis-open.org/cacao/security-playbooks/v2.0/cs01/security-playbooks-v2.0-cs01.html#_Toc152256509) in the spec).

Every action step in a CACAO playbook must have a single agent, and one or more targets. Both agents and targets are defined on the playbook level, in the `agent_definitions` and `target_definitions` properties. SOARCA will execute action steps that have an agent of the type `soarca`. The capability that will be selected to execute the step is determined by the `name` property of the agent. For more information, [read the documentation on components](/docs/core-components/modules).

```json
"target_definitions": {
    "linux--b49069c2-0b69-4a46-8509-80196c4a9bf8": {
      "type": "linux",
      "name": "Target system",
      "description": "System to execute commands on",
      "address": {
        "ipv4": [
          "__target_ip__:value"
        ]
      }
    }
}
```

### Start and end steps

Every CACAO playbook should start with a `start` step. From there, each step can define which step should be executed after the current step finishes. Depending on whether the current step has executed successfully, the next step is defined by the `on_completion` property. A non-successful step execution may instead trigger the `workflow_exception` step, specified in the playbook properties. Alternatively to `on_completion`, a step can specify `on_success` and `on_failure`, which allow a finer-grain control over the execution flow.

What constitutes a successful step execution and what is a failure depends on the specific capability executing the step.

Example of a `start` step:

```json
"start--d6c44626-c9b6-426b-ad5d-3311bafaf068": {
    "on_completion": "action--4b08af84-3741-48ca-8c92-df1557a87379",
    "type": "start"
}
```

According to the CACAO specification, every branch of steps should end in a (unique) `end` step. This is the only step type may not specify a next step:

```json
"end--60fc8d0c-3677-4363-8576-9ea9014f8c8e": {
    "type": "end"
}
```

### If-condition, while-condition, and parallel steps

An `if-condition` step allows executing different branches depending on a specified condition. The step must specify an `on_true` property, which references the start of a branch of steps that should be executed if the condition evaluates to true. Optionally, the `if-condition` step can define an `on_false` property that defines an alternative branch that is executed if the condition evaluates to false. In each case, the specified branch keeps executing until it encounters an `end` step.

The `condition` property contains a string that specifies a [STIX Pattern](https://docs.oasis-open.org/cti/stix/v2.0/stix-v2.0-part5-stix-patterning.html). Currently, SOARCA only supports a very small subsection of the STIX Patterning specification. We support string based equality (`'a' = 'a'`) and inequality (`'a' != 'b'`) comparison. Example:

```json
"if-condition--4b95eaa4-944a-4a9d-88d4-1374a70dbacd": {
    "name": "If it is not new years",
    "description": "Checks if it is 01-01-2025",
    "on_completion": "end--db937fc8-3a42-41cc-b828-ec2db212f425",
    "type": "if-condition",
    "condition": "__soarca_ssh_result__:value != '01-01-2025'",
    "on_true": "action--7fe08053-3685-4d8c-bc0a-40efce75113e"
}
```

SOARCA supports variable interpolation, which means that variables can be used inside the `condition` property, as seen in the example above.

Similarly, CACAO specifies `while-condition` steps, whose `on_true` branch will be repeatedly executed until the condition evaluates to false.

The `parallel` step allows executing multiple branches (in parallel) specified in the `next_steps` property. At this time, the steps in `next_steps` are executed sequentially. Parallel execution is scheduled for a later release in SOARCA.

Next, we explain variables in CACAO and SOARCA.

### Variables

The CACAO specification allows defining variables on the playbook level, as well as on the step level. Playbook variables are available throughout the playbook. In SOARCA, variables defined on the step level are available in that step, and in any step that executes in a sub-branch of an `if-condition`, `while-condition`, or `parallel` step.

According to the CACAO spec, variable names should start and end with double underscores (`__`). The CACAO spec allows defining multiple types of variables (strings, ip-addresses, numbers), but at this time SOARCA will interpret every value as a string. The `constant` and `external` properties are ignored.

```json
"playbook_variables": {
    "__target_ip__": {
      "type": "ipv4-addr",
      "description": "IP address of target system",
      "constant": false,
      "external": true
    }
}
```

SOARCA supports the interpolation of variables in different strings. The specific string-based properties that support interpolation depend on the capability. In general, string interpolation is supported in the properties of agents, targets, authentication information, and `command` properties.

Variable interpolation happens at the last possible moment, which means that step-dependant variables can be used in agent and target definitions.

Substitution is performed by replacing any occurrence of `[variable_name]:value` with the string `value` of that variable. Undefined variables are not replaced.

### Action steps

Within CACAO playbooks, `action` steps can define commands that are executed by an _agent_ against one or more _targets_. The agent and targets are referenced by ID. SOARCA selects the internal capability for handling the step by looking at the `type` and `name` of the agent. After selecting the proper capability, SOARCA will sequentially perform every command in the `commands` property for every target specified in `targets`. If any command fails to execute successfully, further execution is halted and the step is considered to have failed.

Action steps may return variables. On the successful execution of an action step, any variables returned are added to the globally available playbook variables. If the `out_args` property is specified and non-empty, only the variables listed in `out_args` will be added to the global playbook variables. The `in_args` property from the CACAO spec is ignored. Any variable defined on the playbook level, in parent-steps and within the step itself are available for interpolation.

In the case an action step ends in a failure, any variables returned from the step are ignored.

The example below shows how to run an `ssh` command on a single target system:

```json
"action--4b08af84-3741-48ca-8c92-df1557a87379": {
    "name": "Get current time current system",
    "step_variables": {
        "__soarca_ssh_result__": {
            "type": "string",
            "description": "Output of the ssh command",
            "constant": false,
            "external": false
        }
    },
    "on_completion": "if-condition--4b95eaa4-944a-4a9d-88d4-1374a70dbacd",
    "type": "action",
    "commands": [
        {
            "type": "ssh",
            "description": "Retrieve date",
            "command": "date -I | tr -d \"\\n\""
        }
    ],
    "agent": "soarca--664bbe4a-7ad3-462c-baca-53cee8d67594",
    "targets": [
        "linux--b49069c2-0b69-4a46-8509-80196c4a9bf8"
    ],
    "out_args": [
        "__soarca_ssh_result__"
    ]
},
```

## Example playbook

This is de JSON data of the playbook used throughout this page.

```json
{
  "type": "playbook",
  "spec_version": "cacao-2.0",
  "id": "playbook--52f8cd0d-179a-48bf-aa90-32401fe6993c",
  "name": "Example Playbook",
  "description": "Playbook demonstrating SOARCA 1.0 capabilities",
  "playbook_types": [
    "mitigation"
  ],
  "playbook_activities": [
    "step-sequence"
  ],
  "created_by": "identity--dd22fb7f-af84-4957-84ed-12deb6c42d5d",
  "created": "2024-03-07T15:16:19.068Z",
  "modified": "2024-03-07T15:16:19.068Z",
  "revoked": false,
  "derived_from": [
    "playbook--77995581-d375-4905-bd8c-55f820a3e1a3"
  ],
  "playbook_variables": {
    "__target_ip__": {
      "type": "ipv4-addr",
      "description": "IP address of target system",
      "constant": false,
      "external": true
    },
    "__openc2_actuator_ip__": {
      "type": "ipv4-addr",
      "description": "IP address of OpenC2 actuator",
      "constant": false,
      "external": true
    }
  },
  "workflow_start": "start--d6c44626-c9b6-426b-ad5d-3311bafaf068",
  "workflow": {
    "start--d6c44626-c9b6-426b-ad5d-3311bafaf068": {
      "on_completion": "action--4b08af84-3741-48ca-8c92-df1557a87379",
      "type": "start"
    },
    "action--4b08af84-3741-48ca-8c92-df1557a87379": {
      "name": "Get current time current system",
      "step_variables": {
        "__soarca_ssh_result__": {
          "type": "string",
          "description": "Output of the ssh command",
          "constant": false,
          "external": false
        }
      },
      "on_completion": "if-condition--4b95eaa4-944a-4a9d-88d4-1374a70dbacd",
      "type": "action",
      "commands": [
        {
          "type": "ssh",
          "description": "Retrieve date",
          "command": "date -I | tr -d \"\\n\""
        }
      ],
      "agent": "soarca--664bbe4a-7ad3-462c-baca-53cee8d67594",
      "targets": [
        "linux--b49069c2-0b69-4a46-8509-80196c4a9bf8"
      ],
      "out_args": [
        "__soarca_ssh_result__"
      ]
    },
    "if-condition--4b95eaa4-944a-4a9d-88d4-1374a70dbacd": {
      "name": "If it is not new years",
      "description": "Checks if it is 01-01-2025",
      "on_completion": "end--db937fc8-3a42-41cc-b828-ec2db212f425",
      "type": "if-condition",
      "condition": "__soarca_ssh_result__:value != '01-01-2025'",
      "on_true": "action--7fe08053-3685-4d8c-bc0a-40efce75113e"
    },
    "action--7fe08053-3685-4d8c-bc0a-40efce75113e": {
      "name": "Perform an HTTP request",
      "description": "Perform a GET request against httpbin.org",
      "step_variables": {
        "__soarca_http_api_result__": {
          "type": "string",
          "constant": false,
          "external": false
        }
      },
      "on_completion": "while-condition--d865da4e-4f53-4b29-aaba-b8f5711d50ff",
      "type": "action",
      "commands": [
        {
          "type": "http-api",
          "description": "Perform request against httpbin.org",
          "command": "GET /get?newyears=false HTTP/1.1"
        }
      ],
      "targets": [
        "http-api--f2ed7db1-54fc-4a3c-ac0c-837dffade754"
      ],
      "out_args": [
        "__soarca_http_api_result__"
      ]
    },
    "while-condition--d865da4e-4f53-4b29-aaba-b8f5711d50ff": {
      "name": "While the counter is not 5",
      "description": "Step showing while condition",
      "step_variables": {
        "__soarca_ssh_result__": {
          "type": "string",
          "description": "Incrementing counter",
          "value": "0",
          "constant": false,
          "external": false
        }
      },
      "on_completion": "action--76fe4c02-6a5d-43ae-8736-433c07ab80b8",
      "type": "while-condition",
      "condition": "__soarca_ssh_result__:value != '5'",
      "on_true": "action--a32cdbb6-403a-47c7-a35b-07430a8de3fd"
    },
    "action--a32cdbb6-403a-47c7-a35b-07430a8de3fd": {
      "name": "Increment the counter",
      "description": "Increment the counter stored in __soarca_ssh_result__",
      "step_variables": {
        "__soarca_ssh_result__": {
          "type": "string",
          "constant": false,
          "external": false
        }
      },
      "on_completion": "end--60fc8d0c-3677-4363-8576-9ea9014f8c8e",
      "type": "action",
      "commands": [
        {
          "type": "ssh",
          "description": "Increment a string counter using python",
          "command": "python -c \"print(__soarca_ssh_result__:value + 1, end='')\""
        }
      ],
      "targets": [
        "linux--b49069c2-0b69-4a46-8509-80196c4a9bf8"
      ],
      "out_args": [
        "__soarca_ssh_result__"
      ]
    },
    "end--60fc8d0c-3677-4363-8576-9ea9014f8c8e": {
      "type": "end"
    },
    "end--db937fc8-3a42-41cc-b828-ec2db212f425": {
      "type": "end"
    },
    "action--76fe4c02-6a5d-43ae-8736-433c07ab80b8": {
      "name": "Send OpenC2 command",
      "description": "Sends a command to an OpenC2 compliant actuator",
      "step_variables": {
        "__soarca_openc2_http_result__": {
          "type": "string",
          "constant": false,
          "external": false
        }
      },
      "on_completion": "end--40f92845-e67a-4f13-b72a-23f189bf0cb6",
      "type": "action",
      "commands": [
        {
          "type": "openc2-http",
          "command": "POST /openc2-api/ HTTP/1.1",
          "content_b64": "ewogICJoZWFkZXJzIjogewogICAgInJlcXVlc3RfaWQiOiAiZDFhYzA0ODktZWQ1MS00MzQ1LTkxNzUtZjMwNzhmMzBhZmU1IiwKICAgICJjcmVhdGVkIjogMTU0NTI1NzcwMDAwMCwKICAgICJmcm9tIjogInNvYXJjYS5ydW5uZXIubmV0IiwKICAgICJ0byI6IFsKICAgICAgImZpcmV3YWxsLmFwaS5jb20iCiAgICBdCiAgfSwKICAiYm9keSI6IHsKICAgICJvcGVuYzIiOiB7CiAgICAgICJyZXF1ZXN0IjogewogICAgICAgICJhY3Rpb24iOiAiZGVueSIsCiAgICAgICAgInRhcmdldCI6IHsKICAgICAgICAgICJmaWxlIjogewogICAgICAgICAgICAiaGFzaGVzIjogewogICAgICAgICAgICAgICJzaGEyNTYiOiAiMjJmZTcyYTM0ZjAwNmVhNjdkMjZiYjcwMDRlMmI2OTQxYjVjMzk1M2Q0M2FlN2VjMjRkNDFiMWE5MjhhNjk3MyIKICAgICAgICAgICAgfQogICAgICAgICAgfQogICAgICAgIH0KICAgICAgfQogICAgfQogIH0KfQ==",
          "headers": {
            "Content-Type": ["application/openc2+json;version=1.0"]
          }
        }
      ],
      "targets": [
        "linux--c50901f4-3802-4b8f-9b19-e1e62cc8fac4"
      ],
      "out_args": [
        "__soarca_openc2_http_result__"
      ]
    },
    "end--40f92845-e67a-4f13-b72a-23f189bf0cb6": {
      "type": "end"
    }
  },
  "agent_definitions": {
    "soarca--664bbe4a-7ad3-462c-baca-53cee8d67594": {
      "type": "soarca",
      "name": "soarca-ssh",
      "description": "SOARCA SSH capability"
    }
  },
  "target_definitions": {
    "linux--b49069c2-0b69-4a46-8509-80196c4a9bf8": {
      "type": "linux",
      "name": "Target system",
      "description": "System to execute commands on",
      "address": {
        "ipv4": [
          "__target_ip__:value"
        ]
      }
    },
    "http-api--f2ed7db1-54fc-4a3c-ac0c-837dffade754": {
      "type": "http-api",
      "name": "HTTPBin",
      "description": "The HTTPBin.org testing website",
      "address": {
        "url": [
          "https://httpbin.org/"
        ]
      }
    },
    "linux--c50901f4-3802-4b8f-9b19-e1e62cc8fac4": {
      "type": "http-api",
      "name": "OpenC2 Actuator",
      "description": "OpenC2 compatiable actuator",
      "address": {
        "ipv4": [
          "__openc2_actuator_ip__:value"
        ]
      }
    }
  }
}
```