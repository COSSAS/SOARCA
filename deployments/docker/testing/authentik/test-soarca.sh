#!/usr/bin/env bash
# Checks SOARCA's permissions with real authentik tokens.
#
# Run it with authentik up (docker compose up -d in this folder) and SOARCA
# running with soarca-env.sh:
#
#   ./deployments/docker/testing/authentik/test-soarca.sh
#
# Needs curl and jq. SOARCA_URL and AUTHENTIK_URL override the defaults.
set -euo pipefail

AUTHENTIK=${AUTHENTIK_URL:-http://localhost:9000}
SOARCA=${SOARCA_URL:-http://localhost:8080}
CLIENT_ID=${SOARCA_CLIENT_ID:-soarca}
passed=0
failed=0

token() { # username; the app password is "<username>-token"
  curl -sf "$AUTHENTIK/application/o/token/" \
    -d grant_type=client_credentials \
    --data-urlencode "client_id=$CLIENT_ID" \
    --data-urlencode "username=$1" \
    --data-urlencode "password=$1-token" \
    -d "scope=openid profile email soarca" | jq -r .access_token
}

claims() { # decode a JWT payload
  local payload
  payload=$(printf '%s' "$1" | cut -d. -f2 | tr '_-' '/+')
  while [ $((${#payload} % 4)) -ne 0 ]; do payload="$payload="; done
  printf '%s' "$payload" | base64 -d
}

# check NAME EXPECTED CURL_ARGS...
# EXPECTED is a status code, or "allowed" for anything except 401 and 403:
# then the request got past authentication and authorization, and whatever
# the handler answers (200, 400, 404) is fine.
check() {
  local name=$1 expected=$2
  shift 2
  local status
  status=$(curl -s -o /tmp/soarca-authentik-body -w '%{http_code}' "$@")
  local ok=false
  if [ "$expected" = allowed ]; then
    [ "$status" != 401 ] && [ "$status" != 403 ] && [ "$status" != 000 ] && ok=true
  else
    [ "$status" = "$expected" ] && ok=true
  fi
  if $ok; then
    passed=$((passed + 1))
    printf '  ok    %-52s %s\n' "$name" "$status"
  else
    failed=$((failed + 1))
    printf '  FAIL  %-52s got %s, want %s: %s\n' "$name" "$status" "$expected" "$(head -c 200 /tmp/soarca-authentik-body)"
  fi
}

curl -sf -o /dev/null "$SOARCA/status/ping" || { echo "SOARCA isn't answering on $SOARCA" >&2; exit 1; }

echo "Requesting tokens from $AUTHENTIK"
admin=$(token soc-admin)
author=$(token author)
analyst=$(token analyst)
reader=$(token reader)
webhook=$(token svc-webhook)
for value in "$admin" "$author" "$analyst" "$reader" "$webhook"; do
  if [ -z "$value" ] || [ "$value" = null ]; then
    echo "Could not get a token. Is authentik up and the blueprint applied? (docker compose logs blueprint)" >&2
    exit 1
  fi
done
echo "Analyst token:"
claims "$analyst" | jq '{iss, aud, groups, group_ids, exp}'
echo

bearer() { printf 'Authorization: Bearer %s' "$1"; }
json='Content-Type: application/json'
playbook='{"type":"playbook","spec_version":"cacao-2.0","id":"playbook--00000000-0000-4000-8000-000000000422","name":"authentik test","created_by":"identity--00000000-0000-4000-8000-000000000001","created":"2026-01-01T00:00:00Z","modified":"2026-01-01T00:00:00Z","workflow_start":"start--00000000-0000-4000-8000-000000000001","workflow":{"start--00000000-0000-4000-8000-000000000001":{"type":"start","on_completion":"end--00000000-0000-4000-8000-000000000001"},"end--00000000-0000-4000-8000-000000000001":{"type":"end"}}}'

echo "Without a valid token"
check "ping needs no token"                         200 "$SOARCA/status/ping"
check "playbooks need a token"                      401 "$SOARCA/playbook/"
check "a tampered token is rejected"                401 -H "$(bearer "${reader}x")" "$SOARCA/playbook/"

echo "reader (soarca_read)"
check "reads playbooks"                             200 -H "$(bearer "$reader")" "$SOARCA/playbook/"
check "reads status"                                200 -H "$(bearer "$reader")" "$SOARCA/status/"
check "cannot create playbooks"                     403 -X POST -H "$(bearer "$reader")" -H "$json" -d "$playbook" "$SOARCA/playbook/"
check "cannot trigger playbooks"                    403 -X POST -H "$(bearer "$reader")" "$SOARCA/trigger/playbook/unknown"

echo "author (soarca_write)"
check "creates a playbook"                          allowed -X POST -H "$(bearer "$author")" -H "$json" -d "$playbook" "$SOARCA/playbook/"
check "cannot trigger playbooks"                    403 -X POST -H "$(bearer "$author")" "$SOARCA/trigger/playbook/unknown"

echo "analyst (soarca_execute)"
check "reads playbooks"                             200 -H "$(bearer "$analyst")" "$SOARCA/playbook/"
check "passes the trigger check"                    allowed -X POST -H "$(bearer "$analyst")" "$SOARCA/trigger/playbook/unknown"
check "cannot delete playbooks"                     403 -X DELETE -H "$(bearer "$analyst")" "$SOARCA/playbook/playbook--00000000-0000-4000-8000-000000000422"

echo "svc-webhook (service account, soarca_execute)"
check "passes the trigger check"                    allowed -X POST -H "$(bearer "$webhook")" "$SOARCA/trigger/playbook/unknown"
check "cannot reach key management"                 403 -H "$(bearer "$webhook")" "$SOARCA/keymanagement/"

echo "soc-admin (soarca_admin)"
check "deletes the test playbook"                   allowed -X DELETE -H "$(bearer "$admin")" "$SOARCA/playbook/playbook--00000000-0000-4000-8000-000000000422"
check "passes the trigger check"                    allowed -X POST -H "$(bearer "$admin")" "$SOARCA/trigger/playbook/unknown"

echo
echo "$passed passed, $failed failed"
[ "$failed" = 0 ]
