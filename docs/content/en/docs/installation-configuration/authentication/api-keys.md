---
title: API keys
description: >
  Static keys for systems that cannot request an OAuth token
categories: [extensions, architecture]
tags: [security]
weight: 3
---

Some systems can only send a fixed header, for example a webhook source that triggers a playbook. For those, SOARCA accepts static API keys. Prefer [Entra ID](../entra/) or [authentik](../authentik/) wherever the caller can request a token: tokens expire within hours, while a leaked key stays valid until it's removed or expires.

SOARCA never stores the keys themselves, only their SHA-256 hash. Every key starts with `soarca_`, which lets SOARCA recognise it and lets secret scanners find leaked ones.

## Create a key

```bash
KEY="soarca_$(openssl rand -base64 32 | tr '+/' '-_' | tr -d '=')"
HASH=$(printf '%s' "$KEY" | sha256sum | cut -d' ' -f1)
echo "Key for the caller: $KEY"
echo "Hash for SOARCA:    $HASH"
```

Give the key to the caller once, through a secure channel, and don't keep it anywhere else. On macOS, use `shasum -a 256` instead of `sha256sum`.

## Register it with SOARCA

Add an entry to the API keys file, a JSON array:

```json
[
  {
    "name": "webhook",
    "sha256": "<HASH>",
    "permissions": ["playbook:execute"],
    "expires": "2027-01-31T00:00:00Z"
  }
]
```

| Field | Description |
|---|---|
| `name` | Identifies the caller in logs. Use a different name for every key. |
| `sha256` | The hex SHA-256 hash of the key. |
| `permissions` | The [permissions](../#permissions) the key grants. A key grants exactly what's listed; `playbook:execute` doesn't include `playbook:read` here. |
| `expires` | When the key stops working, in RFC 3339 format. Required. |

Point SOARCA to the file:

{{< tabpane langEqualsHeader=false >}}
{{< tab header="`.env`" lang="txt" >}}
AUTH_ENABLED: true
AUTH_API_KEYS_FILE: "/etc/soarca/api-keys.json"
{{< /tab >}}
{{< /tabpane >}}

SOARCA reads the file at startup and refuses to start if an entry has no name, an invalid hash or no expiry. Restart SOARCA after changing the file. Keep the file readable only by the user that runs SOARCA.

## Use it

```bash
curl -X POST -H "Authorization: Bearer $KEY" -H "Content-Type: application/json" \
  -d @playbook.json http://localhost:8080/trigger/playbook
```

## Revoke or rotate a key

Remove the entry from the file, or add a new entry first and remove the old one once the caller has switched, then restart SOARCA.
