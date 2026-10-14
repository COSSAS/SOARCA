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

SOARCA grants access through four permissions. Each identity provider maps its own roles or groups onto them; see the provider pages below for the exact names.

| Permission | Allows |
|---|---|
| `playbook:read` | Reading playbooks, execution reports and status |
| `playbook:write` | Creating, updating and deleting playbooks |
| `playbook:execute` | Triggering playbooks and answering manual steps |
| `soarca:admin` | Everything, including key management |

Every role that grants `playbook:write` or `playbook:execute` also grants `playbook:read`.

| Route | Required permission |
|---|---|
| `GET /playbook/`, `GET /playbook/meta/`, `GET /playbook/:id` | `playbook:read` |
| `POST /playbook/`, `PUT /playbook/:id`, `DELETE /playbook/:id` | `playbook:write` |
| `GET /reporter/`, `GET /reporter/:id`, `GET /status/`, `GET /status/ping` | `playbook:read` |
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

### OpenID Connect (authentik)

| Variable | Default | Description |
|---|---|---|
| `OIDC_ISSUER` | | Issuer URL, exactly as it appears in the tokens' `iss` claim, including any trailing slash. |
| `OIDC_CLIENT_ID` | | Client ID of the SOARCA provider. Tokens must carry it as their audience. |
| `OIDC_SKIP_TLS_VERIFY` | `false` | Skip certificate checks towards the provider. Only for test setups with self-signed certificates. |
| `AUTH_GROUP` | `soarca_admin` | Members of this group get every permission. |

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
