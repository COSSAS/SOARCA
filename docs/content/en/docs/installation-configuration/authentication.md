
---

title: Setup RBAC for SOARCA
description: >
  Setup Role-Based Access Control (RBAC) for SOARCA
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

- Client ID': `some random stuff`
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
OIDC_ISSUER: "<https://authentikuri:9443/application/o/soarca/>>"
OIDC_CLIENT_ID: "WxUcBMGZdI7c0e5oYp6mYdEd64acpXSuWKh8zBH5"
OIDC_SKIP_TLS_VERIFY: true
{{< /tab >}}
{{< /tabpane >}}
