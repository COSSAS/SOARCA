# SOARCA settings for the authentik test setup in this folder. Load them before
# starting SOARCA from the repository root:
#
#   set -a; . deployments/docker/testing/authentik/soarca-env.sh; set +a
#   ./build/soarca
AUTH_ENABLED=true
OIDC_ISSUER=http://localhost:9000/application/o/soarca/
OIDC_CLIENT_ID=soarca
OIDC_SKIP_TLS_VERIFY=false
AUTH_GROUP=soarca_admin
