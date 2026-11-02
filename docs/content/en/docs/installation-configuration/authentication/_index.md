---
title: Authentication and authorization
description: >
  Secure the SOARCA API with Microsoft Entra ID, authentik or another OIDC provider, and static API keys
categories: [extensions, architecture]
tags: [security]
weight: 2
---

Authentication is disabled by default in SOARCA. With the default configuration, anyone who can reach SOARCA over the network can submit and execute playbooks. Because playbooks can reconfigure the systems SOARCA is connected to, enable authentication for every deployment that is reachable by more than a single trusted machine.

With authentication enabled, every API request needs an `Authorization: Bearer <credential>` header. SOARCA accepts three kinds of credentials, side by side:

| Credential | Issued by | Typical caller |
|---|---|---|
| Entra ID access token | Microsoft Entra ID | Analysts with a Microsoft account, Azure workloads, integrations registered in Entra |
| OIDC access token | [authentik](https://goauthentik.io/) or another OpenID Connect provider | Analysts and service accounts managed in authentik |
| API key | SOARCA itself | Systems that cannot request an OAuth token, such as webhook sources |

You can enable any combination. SOARCA looks at each credential and hands it to the matching provider: a JWT by its issuer (`iss` claim), an API key by its `soarca_` prefix. Tokens from an issuer SOARCA isn't configured for are rejected.

## Permissions

SOARCA grants access through five permissions. Each identity provider maps its own roles or groups onto them; see the provider pages below for the exact names.

| Permission | Allows |
|---|---|
| `playbook:read` | Reading playbooks and execution reports |
| `playbook:write` | Creating, updating and deleting playbooks |
| `playbook:execute` | Triggering playbooks and answering manual steps |
| `status:read` | Read SOARCA status API |
| `soarca:admin` | Everything, including key management |

Every role that grants `playbook:write` or `playbook:execute` also grants `playbook:read`.

| Route | Required permission |
|---|---|
| `GET /playbook/`, `GET /playbook/meta/`, `GET /playbook/:id` | `playbook:read` |
| `POST /playbook/`, `PUT /playbook/:id`, `DELETE /playbook/:id` | `playbook:write` |
| `GET /reporter/`, `GET /reporter/:id`, `GET /principal/permissions` | `playbook:read` |
| `GET /status/` | `status:read` |
| `GET /status/ping`, `GET /.well-known/oauth-protected-resource` | None (public) |
| `POST /trigger/playbook`, `POST /trigger/playbook/:id` | `playbook:execute` |
| `GET /manual/`, `GET /manual/:exec_id/:step_id`, `POST /manual/continue` | `playbook:execute` |
| `/keymanagement/*` | `soarca:admin` |

A request without a valid credential gets `401 Unauthorized`. A valid credential that lacks the permission gets `403 Forbidden`. Both use SOARCA's usual error format:

```json
{
  "status": "403",
  "message": "forbidden: requires playbook:execute",
  "original-call": "POST /trigger/playbook",
  "downstream-call": ""
}
```

## Discovering the accepted identity providers

SOARCA publishes which identity providers it accepts as OAuth protected resource metadata ([RFC 9728](https://www.rfc-editor.org/rfc/rfc9728)). The document is public, so clients can read it before anyone signs in:

```bash
curl https://soarca.example.com/.well-known/oauth-protected-resource
```

```json
{
  "resource": "https://soarca.example.com",
  "resource_name": "SOARCA",
  "authorization_servers": [
    "https://authentik.example.com/application/o/soarca/",
    "https://login.microsoftonline.com/11111111-1111-1111-1111-111111111111/v2.0"
  ],
  "scopes_supported": ["api://soarca/access_as_user"],
  "bearer_methods_supported": ["header"]
}
```

`authorization_servers` lists the issuer of every token SOARCA accepts: the OIDC issuer, one Entra ID issuer per tenant, and with `AUTH_ENTRA_ACCEPT_V1` also the v1 issuers. API keys aren't listed, since nobody signs in with them. With `AUTH_ENABLED: false` the list is empty, which tells clients they can skip sign-in. `scopes_supported` holds the Entra ID scope that user tokens must carry, when Entra ID is enabled.

A client such as SOARCA-GUI keeps its own client registration per provider. It offers sign-in only for providers whose issuer appears in `authorization_servers`, and can warn at startup when a provider it is configured for isn't listed, instead of every call failing with `401` after sign-in.

A `401` response points at the document in its `WWW-Authenticate` header, so a client can discover it from a failed call:

```
WWW-Authenticate: Bearer realm="soarca", resource_metadata="https://soarca.example.com/.well-known/oauth-protected-resource"
```

`resource` is taken from the host the request came in on. Behind a reverse proxy, set `AUTH_RESOURCE_URL` to SOARCA's public URL.

{{% alert title="Public information" color="info" %}}
The document contains your Entra ID tenant IDs. They aren't secret (Entra ID publishes them through its own discovery documents), but anyone who can reach SOARCA can read them.
{{% /alert %}}

## Checking what a caller may do

`GET /principal/permissions` returns the permissions of whoever calls it: a user, an application or an API key. Clients such as SOARCA-GUI use it to hide actions the caller can't perform. SOARCA still checks the permission on every call.

```json
{
  "subject": "entra:11111111-1111-1111-1111-111111111111:user:3f2c…",
  "provider": "entra",
  "permissions": ["playbook:read", "playbook:execute"]
}
```

`soarca:admin` is expanded, so an administrator gets the full list and clients don't need to know which permission implies which.

The endpoint needs `playbook:read`, which every role grants. A caller that signed in but holds no permission gets `403 Forbidden`, which clients can show as "no access to SOARCA". An API key that only holds `playbook:execute` can't call it.

## Setting up a provider

- [Microsoft Entra ID](entra/)
- [authentik](authentik/) (and other OpenID Connect providers)
- [API keys](api-keys/)

## Environment variables

Authentication is turned on with `AUTH_ENABLED: true`. SOARCA then refuses to start unless at least one provider is fully configured.

### General

| Variable | Default | Description |
|---|---|---|
| `AUTH_ENABLED` | `false` | Require a valid credential on every API request. |
| `AUTH_RESOURCE_URL` | | SOARCA's public URL, e.g. `https://soarca.example.com`, published in the [protected resource metadata](#discovering-the-accepted-identity-providers). Empty takes it from each request. Must not contain a path. |

### OpenID Connect (authentik)

| Variable | Default | Description |
|---|---|---|
| `OIDC_ISSUER` | | Issuer URL, exactly as it appears in the tokens' `iss` claim, including any trailing slash. |
| `OIDC_CLIENT_ID` | | Client ID of the SOARCA provider. Tokens must carry it as their audience. |
| `OIDC_SKIP_TLS_VERIFY` | `false` | Skip certificate checks towards the provider. Only for test setups with self-signed certificates. |
| `AUTH_ADMIN_GROUP` | `soarca_admin` | Members of this group get every permission. |

`OIDC_ISSUER` and `OIDC_CLIENT_ID` must be set together.

### Microsoft Entra ID

| Variable | Default | Description |
|---|---|---|
| `AUTH_ENTRA_TENANT_IDS` | | Comma-separated tenant IDs whose tokens SOARCA accepts. |
| `AUTH_ENTRA_API_CLIENT_ID` | | Application (client) ID of the SOARCA API app registration. |
| `AUTH_ENTRA_API_APP_ID_URI` | | Application ID URI of that registration, e.g. `api://soarca`. Optional. |
| `AUTH_ENTRA_ALLOWED_CLIENT_IDS` | | Comma-separated client IDs of the applications allowed to call SOARCA. Empty allows every application in an allowed tenant. |
| `AUTH_ENTRA_REQUIRED_USER_SCOPE` | `access_as_user` | Delegated scope that user tokens must carry. |
| `AUTH_ENTRA_ACCEPT_V1` | `false` | Also accept v1 tokens (issuer `https://sts.windows.net/<tenant>/`). |
| `AUTH_ENTRA_AUTHORITY` | `https://login.microsoftonline.com` | Login host. Change it only for national clouds, e.g. `https://login.microsoftonline.us`. |

`AUTH_ENTRA_TENANT_IDS` and `AUTH_ENTRA_API_CLIENT_ID` must be set together.

### API keys

| Variable | Default | Description |
|---|---|---|
| `AUTH_API_KEYS_FILE` | | Path to a JSON file with hashed API keys. Empty disables API keys. |

### Example

{{< tabpane langEqualsHeader=false >}}
{{< tab header="`.env`" lang="txt" >}}
AUTH_ENABLED: true

# authentik
OIDC_ISSUER: "https://authentik.example.com/application/o/soarca/"
OIDC_CLIENT_ID: "soarca"
AUTH_GROUP: "soarca_admin"

# Entra ID
AUTH_ENTRA_TENANT_IDS: "11111111-1111-1111-1111-111111111111"
AUTH_ENTRA_API_CLIENT_ID: "aaaaaaaa-0000-0000-0000-000000000001"
AUTH_ENTRA_API_APP_ID_URI: "api://soarca"

# API keys
AUTH_API_KEYS_FILE: "/etc/soarca/api-keys.json"
{{< /tab >}}
{{< /tabpane >}}

## Checking a token

Access tokens are signed, not encrypted, so you can read their claims. This is the quickest way to find out why SOARCA rejects one:

```bash
echo "$TOKEN" | cut -d. -f2 | tr '_-' '/+' | base64 -d 2>/dev/null | jq
```

Compare `iss` with the issuer SOARCA is configured for, `aud` with the client ID, and `roles` (Entra ID) or `groups` (authentik) with the role tables on the provider pages.

Then call SOARCA with it:

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/playbook/
```

## Common problems

| Response | Cause |
|---|---|
| `401` for every request | The token's issuer isn't configured, the token expired, or it was issued for another audience. Decode the token and compare `iss` and `aud`. |
| `401` with an Entra token whose `aud` is `https://graph.microsoft.com` | The caller requested a Microsoft Graph token. Request a token for the SOARCA API instead; see [Entra ID](entra/). |
| `403 forbidden: requires …` | The credential is valid, but none of its roles or groups grants that permission. |
| SOARCA doesn't start: `AUTH_ENABLED=true requires …` | Authentication is enabled without a complete provider configuration. |
