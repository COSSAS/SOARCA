
---

title: Setup RBAC
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
AUTH_ENABLED: false
OIDC_ISSUER: "<https://localhost:9443/application/u/test/>"
OIDC_CLIENT_SECRET: "SOME_CLIENT_SECRET"
OIDC_CLIENT_ID: "SOME_CLIENT_ID"
OIDC_REDIRECT_URL: "<http://localhost:8081/auth/soarca_gui/callback>"
COOKIE_SECRET_KEY: "SOME_COOKIE_SECRET" # OPTIONAL: openssl rand -base64 32  or head -c 32 /dev/urandom | base64
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

### Making an application

### Setting the authentication provider
