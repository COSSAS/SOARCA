
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

Currently, for the used [gauth library](https://github.com/COSSAS/gauth) the [Authentik](https://goauthentik.io/) authentication provider, an open-source solution that supports a wide range of authentication methods is supported and tested. As such other OIDC-based providers might not be compatible.

## Supported OIDC-Based Auth Providers

- [Authentik](https://goauthentik.io/)

## Enabling RBAC

Enabling RBAC can be done by setting the `AUTH_ENABLED: true`.

{{< tabpane langEqualsHeader=false  >}}
{{< tab header="`.env`" lang="txt" >}}
AUTH_ENABLED: true
AUTH_GROUP: "soarca_admin"
OIDC_ISSUER: "https://authentikuri:9443/application/u/test/"
OIDC_CLIENT_SECRET: "SOME_CLIENT_SECRET"
OIDC_CLIENT_ID: "SOME_CLIENT_ID"
OIDC_SKIP_TLS_VERIFY: true
{{< /tab >}}
{{< /tabpane >}}

## Setting up Authentik with SOARCA

Next, we need to obtain variables such as `OIDC_ISSUER` etc. This section will describe how we can setup [Authentik](https://goauthentik.io/).

### Spinning up Authentik

Instruction and docker-compose on how to bundle SOARCA with Authentik will come!

### Making an authentication provider

In Authentik first setup a new provider. This can be done under `Applications` -> `Providers` ->`Create`. For the provider type select the `OAuth2/OpenID provider` from the various options. An example configuration is given below:

{{% alert title="Note" color="primary" %}}
We use an token lifetime of **8 hours**, since the [SOARCA-GUI](https://github.com/COSSAS/SOARCA-GUI) uses this token for client validation. As we do not want the user to login every so minute. It is advised to set this to 8 hours.
{{% /alert %}}

Next, we need to set in the advanced protocol settings the token expiration lifetime needs to be changed to **8 hours**.

![core](/SOARCA/images/installation_configuration/authentik_setup/setup-provider.png)

![core](/SOARCA/images/installation_configuration/authentik_setup/change-lifetime.png)

Endpoints for the auth provider can also be found here:

![core](/SOARCA/images/installation_configuration/authentik_setup/endpoints.png)

### Making an application

Next, we can create a new application as shown in the picture below. A new application can be added under `Applications` --> `Create`  Add the earlier made provider to this application.

![core](/SOARCA/images/installation_configuration/authentik_setup/setting-application.png)

### Setting the authentication provider

Next, under `providers` -> `soarca-auth-provider` -> `edit` we can find the following overview:

![core](/SOARCA/images/installation_configuration/authentik_setup/view-provider.png)

Here we can find the:

- Client ID: `some random stuff`
- Client Secret: `some other random stuff`
- Redirect Url: Optional, should be set when using for the SOARCA-GUI explained [here]

![core](/SOARCA/images/installation_configuration/authentik_setup/editing-provider.png)

We only use the Authentik integration for token validation on the SOARCA side. As such, only the `OIDC_CLIENT_ID` is required here. For the SOARCA-GUI, we would need the `OIDC_CLIENT_SECRET`.

{{% alert title="Warning" color="warning" %}}
It is not advised to run Authentik like this! Please setup TLS certificates in a real environment and set the `OIDC_SKIP_TLS_VERIFY` to `false`.
{{% /alert %}}

Set these variables in the environment variables settings, for example:
{{< tabpane langEqualsHeader=false  >}}
{{< tab header="`.env`" lang="txt" >}}
AUTH_ENABLED: true  
AUTH_GROUP: "soarca_admin"
OIDC_ISSUER: "https://authentikuri/application/o/does-providing-for-soarca/"
OIDC_CLIENT_ID: "WxUcBMGZdI7c0e5oYp6mYdEd64acpXSuWKh8zBH5"
OIDC_SKIP_TLS_VERIFY: true
{{< /tab >}}

{{< tab header="`bash`" lang="bash" >}}
export AUTH_ENABLED=true
export AUTH_GROUP="soarca_admin"
export OIDC_ISSUER="https://authentikuri:9443/application/o/does-providing-for-soarca/"
export OIDC_CLIENT_ID="WxUcBMGZdI7c0e5oYp6mYdEd64acpXSuWKh8zBH5"
export OIDC_SKIP_TLS_VERIFY=true
{{< /tab >}}
{{< /tabpane >}}

### Adding SOARCA user group and users

{{% alert title="Note" color="primary" %}}
Again, for the current version of the implementation we only support one group to differentiate between access to the different endpoint. We plan for a later version of SOARCA to have different groups/permissions for a given API endpoint.
{{% /alert %}}

For the current version of SOARCA and the gauth library the access to the API for a given user is dependent on the required set `AUTH_GROUP`. Users are required to be in the same group as the group that has been set through this variable. Currently, there is not yet control over which group can access a specific API or route grooup. This feature is scheduled for a future version of SOARCA. In the example below, the `AUTH_GROUP: soarca_admin` is set.

Next, we require to setup a group in Authentik that is called `soarca_admin` as explained earlier. The to be obtained tokens from Authentik needs to have this group information as this will be checked by the middleware. A group can be created under `Directory` -> `Groups` -> `New`.

Under `users` normal as as service accounts can be created. We advise for machine-to-machine implementation service accounts, and for normal users (used for example for SOARCA-GUI logins) normal accounts. Now we can make an users and add to the `soarca_admin` group. Make use under the application that this group is added to the application provider that we have setup earlier, otherwise the grants of token might fail.

![core](/SOARCA/images/installation_configuration/authentik_setup/groups.png)

![core](/SOARCA/images/installation_configuration/authentik_setup/add-user.png)

![core](/SOARCA/images/installation_configuration/authentik_setup/add-groups.png)

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

{{% alert title="Tip" %}}
For obtaining an access (bearer) token for Authentik, we have provided an example [here](https://github.com/COSSAS/gauth/examples/m2m)
{{% /alert %}}

An example curl command is provided below:

{{< tabpane langEqualsHeader=false  >}}
{{< tab header="`bash`" lang="bash" >}}
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer <replace token>" -d @./example/ssh-playbook.json localhost:8080/trigger/playbook
{{< /tab >}}
{{< /tabpane >}}
