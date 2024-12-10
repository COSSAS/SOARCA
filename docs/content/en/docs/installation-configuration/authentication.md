
---

title: Setup RBAC for SOARCA
description: >
  Setup OIDC based Role-Based Access Control (RBAC) for SOARCA
categories: [extensions, architecture]
tags: [security]
weight: 2
---

Authentication is disabled by default in SOARCA. This means that when SOARCA is launched with its default configuration and exposed to a network, anyone can interact with it. Since SOARCA requires significant capabilities and access to reconfigure systems, exposing it without authentication poses a security risk. This section outlines how to set up authentication and authorization for SOARCA.

SOARCA leverages our internally developed [gauth library](https://github.com/COSSAS/gauth) as its underlying authentication framework. This library provides convenient Role-Based Access Control (RBAC) middleware, which manages authentication for various endpoints, such as the Playbook API. Based on OpenID Connect (OIDC), the library supports integration with multiple authentication providers.

{{% alert title="Warning" color="warning" %}}
Currently, [gauth library](https://github.com/COSSAS/gauth) is supported and tested authentication for the [Authentik](https://goauthentik.io/) authentication provider, an open-source solution that supports a wide range of authentication methods. As such, other OIDC-based providers might not be compatible.
{{% /alert %}}

{{% alert title="Note" color="primary" %}}
For the current version, users/application must belong to the `soarca_admin` group to access the API's when authentication is enabled. Furthermore, currently there is not yet control over what user can go to which route. This feature is scheduled for a future version of SOARCA.
{{% /alert %}}

## Enabling RBAC

Enabling RBAC can be done by setting the `AUTH_ENABLED: true`.

{{< tabpane langEqualsHeader=false  >}}
{{< tab header="`.env`" lang="txt" >}}
AUTH_ENABLED: true
OIDC_ISSUER: "<https://authentikuri:9443/application/u/test/>"
OIDC_CLIENT_SECRET: "SOME_CLIENT_SECRET"
OIDC_CLIENT_ID: "SOME_CLIENT_ID"
OIDC_SKIP_TLS_VERIFY: true
{{< /tab >}}
{{< /tabpane >}}

## Setting up Authentik with SOARCA

Next, we need to obtain variables such as `OIDC_ISSUER` etc. This section will describe how we can setup [Authentik](https://goauthentik.io/).

### Spinning up Authentik

### Making an authentication provider

In Authentik first setup a provider. An example configuration is given below:

![core](/SOARCA/images/installation_configuration/authentik_setup/setup-provider.png)

Next, we need to set in the advanced protocol settings the token expiration lifetime needs to be changed to 8 eight hours.

{{% alert title="Warning" color="warning" %}}
We use an token lifetime of 8 hours, since the [SOARCA-GUI](https://github.com/COSSAS/SOARCA-GUI) uses this token for client validation. As we do not want the user to login every so minute. It is advised to set this to 8 hours.
{{% /alert %}}

![core](/SOARCA/images/installation_configuration/authentik_setup/change-lifetime.png)

Endpoints for the auth provider can also be found here:

![core](/SOARCA/images/installation_configuration/authentik_setup/endpoints.png)

### Making an application

Next create an application as shown in the picture below. Add the earlier made provider to this application.

![core](/SOARCA/images/installation_configuration/authentik_setup/setting-application.png)

### Setting the authentication provider

Next, under `providers` -> `soarca-auth-provider` -> `edit` we can find the following overview:

Here we can find the:

- Client ID: `some random stuff`
- Client Secret': `some other random stuff`
- Redirect Url': Optional, should be set when using for the SOARCA-GUI explained [here]

Set these variables in the environment variables settings, for example:

{{% alert title="Note" color="primary" %}}
We only use the Authentik integration for validation on the SOARCA side. As such, only the `OIDC_CLIENT_ID` is required here. For the SOARCA-GUI, we would need the `OIDC_CLIENT_SECRET`.
{{% /alert %}}

{{% alert title="Warning" color="warning" %}}
It is not advised to run Authentik like this! Please setup TLS certificates in a real environment and set the `OIDC_SKIP_TLS_VERIFY` to `false`.
{{% /alert %}}

{{< tabpane langEqualsHeader=false  >}}
{{< tab header="`.env`" lang="txt" >}}
AUTH_ENABLED: true  
OIDC_ISSUER: "<https://authentikuri/application/o/does-providing-for-soarca/>"
OIDC_CLIENT_ID: "WxUcBMGZdI7c0e5oYp6mYdEd64acpXSuWKh8zBH5"
OIDC_SKIP_TLS_VERIFY: true
{{< /tab >}}

{{< tab header="`bash`" lang="bash" >}}
export AUTH_ENABLED=true
export OIDC_ISSUER="<https://authentikuri:9443/application/o/does-providing-for-soarca/>"
export OIDC_CLIENT_ID="WxUcBMGZdI7c0e5oYp6mYdEd64acpXSuWKh8zBH5"
export OIDC_SKIP_TLS_VERIFY=true
{{< /tab >}}
{{< /tabpane >}}

### Adding SOARCA user group and users

{{% alert title="Note" color="primary" %}}
Again, for the current version of the implementation we only support one group to differentiate between access to the different endpoint. We plan for a later version of SOARCA to have different groups/permissions for a given API endpoint.
{{% /alert %}}

Next, we require to setup a group in Authentik that is called `soarca_admin` as explained earlier. The to be obtained tokens from Authentik needs to have this group information as this will be checked by the middleware.

![core](/SOARCA/images/installation_configuration/authentik_setup/groups.png)

Under `users` normal as as service accounts can be created. We advise for machine-to-machine implementation service accounts, and for normal users (used for example for SOARCA-GUI logins) normal accounts. Now we can make an users and add to the `soarca_admin` group. Make use under the application that this group is added to the application provider that we have setup earlier, otherwise the grants of token might fail.

![core](/SOARCA/images/installation_configuration/authentik_setup/add-user.png)

![core](/SOARCA/images/installation_configuration/authentik_setup/add-group.png)

### Authentication with Bearer

Now that authentication and authorization is enabled, every request requires to have a set `Authorization: Bearer <token>` header.

```
POST /trigger/playbook/ HTTP/1.1
Host: localhost:8080
Content-Type: application/json
Authorization: Bearer <token from authentik> 
Content-Length: 2345

{
    "type": "playbook",
    "spec_version": "cacao-2.0",
    "id": "playbook--300270f9-0e64-42c8-93cc-0927edbe3ae7",
    "name": "Example ssh",
    ...
}

```

The [gauth library](https://github.com/COSSAS/gauth) will validate this bearer token against the setup Authentik provider and grant the user or application access. Replace the token with a working bearer token.  

An example curl command is provided below:
{{< tabpane langEqualsHeader=false  >}}
{{< tab header="`bash`" lang="bash" >}}
curl -X POST "<http://localhost:8080/trigger/playbook/>" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <replace token>" \
-d '{
    "type": "playbook",
    "spec_version": "cacao-2.0",
    "id": "playbook--300270f9-0e64-42c8-93cc-0927edbe3ae7",
    "name": "Example ssh",
    "description": "This playbook demonstrates ssh functionality",
    "playbook_types": ["notification"],
    "created_by": "identity--96abab60-238a-44ff-8962-5806aa60cbce",
    "created": "2023-11-20T15:56:00.123456Z",
    "modified": "2023-11-20T15:56:00.123456Z",
    "valid_from": "2023-11-20T15:56:00.123456Z",
    "valid_until": "2123-11-20T15:56:00.123456Z",
    "priority": 1,
    "severity": 1,
    "impact": 1,
    "labels": ["soarca", "ssh", "example"],
    "authentication_info_definitions": {
        "user-auth--b7ddc2ea-9f6a-4e82-8eaa-be202e942090": {
            "type": "user-auth",
            "username": "root",
            "password": "password"
        }
    },
    "agent_definitions": {
        "soarca--00010001-1000-1000-a000-000100010001": {
            "type": "soarca",
            "name": "soarca-ssh"
        }
    },
    "target_definitions": {
        "ssh--1c3900b4-f86b-430d-b415-12312b9e31f4": {
            "type": "ssh",
            "name": "system 1",
            "address": {
                "ipv4": ["192.168.0.10"]
            },
            "authentication_info": "user-auth--b7ddc2ea-9f6a-4e82-8eaa-be202e942090"
        }
    },
    "external_references": [{
        "name": "TNO COSSAS",
        "description": "TNO COSSAS",
        "source": "TNO COSSAS",
        "url": "https://cossas-project.org"
    }],
    "workflow_start": "start--9e7d62b2-88ac-4656-94e1-dbd4413ba008",
    "workflow_exception": "end--a6f0b81e-affb-4bca-b4f6-a2d5af908958",
    "workflow": {
        "start--9e7d62b2-88ac-4656-94e1-dbd4413ba008": {
            "type": "start",
            "name": "Start ssh example",
            "on_completion": "action--eb9372d4-d524-49fc-bf24-be26ea084779"
        },
        "action--eb9372d4-d524-49fc-bf24-be26ea084779": {
            "type": "action",
            "name": "Execute command",
            "description": "Execute command specified in variable",
            "on_completion": "action--88f4c4df-fa96-44e6-b310-1c06d193ea55",
            "commands": [{
                "type": "ssh",
                "command": "__command__:value"
            }],
            "targets": ["ssh--1c3900b4-f86b-430d-b415-12312b9e31f4"],
            "agent": "soarca--00010001-1000-1000-a000-000100010001",
            "step_variables": {
                "__command__": {
                    "type": "string",
                    "value": "ls -la",
                    "constant": true
                }
            }
        },
        "action--88f4c4df-fa96-44e6-b310-1c06d193ea55": {
            "type": "action",
            "name": "Touch file",
            "description": "Touch file at path specified by variable",
            "on_completion": "end--a6f0b81e-affb-4bca-b4f6-a2d5af908958",
            "commands": [{
                "type": "ssh",
                "command": "touch __path__:value"
            }],
            "targets": ["ssh--1c3900b4-f86b-430d-b415-12312b9e31f4"],
            "agent": "soarca--00010001-1000-1000-a000-000100010001",
            "step_variables": {
                "__path__": {
                    "type": "string",
                    "value": "/root/file1",
                    "constant": true
                }
            }
        },
        "end--a6f0b81e-affb-4bca-b4f6-a2d5af908958": {
            "type": "end",
            "name": "End Flow"
        }
    }
}'
{{< /tab >}}
{{< /tabpane >}}
