# authentik test setup

A local [authentik](https://goauthentik.io/) that comes up configured for SOARCA, for testing OIDC authentication. It uses fixed test passwords and secrets and listens on 127.0.0.1 only, so never run it anywhere else.

## Start

```bash
cd deployments/docker/testing/authentik
docker compose up -d
docker compose wait blueprint        # returns 0 once the SOARCA blueprint is applied
```

The first start takes one to three minutes while authentik migrates its database. If `wait` returns anything but 0, `docker compose logs blueprint` names the blueprint entry that failed.

Then run SOARCA from the repository root with the matching settings:

```bash
make build
set -a; . deployments/docker/testing/authentik/soarca-env.sh; set +a
./build/soarca
```

And check it, in another terminal:

```bash
./deployments/docker/testing/authentik/test-soarca.sh
```

The script requests a token for every test account and checks what each one may do. It needs `curl` and `jq`.

## What's created

| Account | Group | SOARCA permissions |
|---|---|---|
| `soc-admin` | `soarca_admin` | all |
| `author` | `soarca_write` | `playbook:read`, `playbook:write` |
| `analyst` | `soarca_execute` | `playbook:read`, `playbook:execute` |
| `reader` | `soarca_read` | `playbook:read` |
| `svc-webhook` (service account) | `soarca_execute` | `playbook:read`, `playbook:execute` |

- Users sign in to the browser with password `soarca-test`.
- Every account has an app password `<username>-token`, e.g. `analyst-token`, for requesting tokens with curl.
- The OAuth2/OIDC provider has client ID `soarca` and client secret `soarca-test-client-secret`. It allows the authorization code, refresh token and client credentials grants.
- The provider's issuer is `http://localhost:9000/application/o/soarca/`.
- The redirect URI is `http://localhost:8081/auth/soarca_gui/callback`, for SOARCA-GUI.
- The authentik admin UI is at http://localhost:9000/if/admin/, user `akadmin`, password `soarca-admin`.

To change any of these, put the variable from `docker-compose.yml` in a `.env` file next to it and run `docker compose up -d` again.

## Get a token by hand

```bash
TOKEN=$(curl -s http://localhost:9000/application/o/token/ \
  -d grant_type=client_credentials \
  -d client_id=soarca \
  -d username=analyst \
  -d password=analyst-token \
  -d "scope=openid profile email" | jq -r .access_token)

curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/playbook/
```

For the Swagger UI, paste `Bearer <token>` into **Authorize**.

## Things to know

- **Use `localhost:9000` everywhere.** authentik writes the host it was called on into the token's issuer. A token requested through `127.0.0.1:9000` has a different issuer from `OIDC_ISSUER`, and SOARCA answers `401`. For the same reason SOARCA runs on your machine here, not in this compose project: from inside a container, authentik would have another host name.
- **Changing the blueprint.** Edit `blueprints/soarca.yaml` and run `docker compose run --rm blueprint`. Anything the blueprint defines is reset to the file's values. Things you create in the admin UI are left alone.
- **Start over.** `docker compose down -v` deletes authentik's database.
- **Port 9000 in use.** Set `AUTHENTIK_PORT` in `.env`, and change the port in `soarca-env.sh` and in `AUTHENTIK_URL` for the test script.
