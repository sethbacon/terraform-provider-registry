#!/usr/bin/env bash
# fetch-spec.sh — Pull openapi3.json from a pinned backend image.
#
# Strategy:
#   1. Start postgres + backend on a private docker network.
#   2. Wait for the backend's /openapi3.json endpoint (served from embedded
#      bytes once the HTTP server starts after a successful DB ping).
#   3. curl the spec, write it next to this script.
#   4. Tear down both containers and the network.
#
# To bump the backend pin, edit BACKEND_TAG below. CI re-fetches the spec
# on every PR and fails on `git diff`, so the pin and the committed
# openapi3.json stay in lockstep.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_TAG="${BACKEND_TAG:-1.1.6}"
IMAGE="ghcr.io/sethbacon/terraform-registry-backend:${BACKEND_TAG}"
SUFFIX="${RANDOM}-$$"
NETWORK="tpr-spec-fetch-${SUFFIX}"
PG_CONTAINER="tpr-spec-fetch-pg-${SUFFIX}"
BE_CONTAINER="tpr-spec-fetch-be-${SUFFIX}"
PORT=8088
OUT="${SCRIPT_DIR}/openapi3.json"

cleanup() {
  docker rm -f "${BE_CONTAINER}" >/dev/null 2>&1 || true
  docker rm -f "${PG_CONTAINER}" >/dev/null 2>&1 || true
  docker network rm "${NETWORK}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "    image:  ${IMAGE}"
echo "    output: ${OUT}"

docker network create "${NETWORK}" >/dev/null

# Postgres for the backend to ping at startup.
docker run -d \
  --name "${PG_CONTAINER}" \
  --network "${NETWORK}" \
  -e POSTGRES_DB=terraform_registry \
  -e POSTGRES_USER=registry \
  -e POSTGRES_PASSWORD=registry \
  postgres:16-alpine >/dev/null

# Wait for postgres to accept connections.
for i in $(seq 1 30); do
  if docker exec "${PG_CONTAINER}" pg_isready -U registry -d terraform_registry >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

# Start the backend. ENCRYPTION_KEY + TFR_JWT_SECRET are required at boot
# but never used — we hit /openapi3.json which is served from embedded bytes.
docker run -d \
  --name "${BE_CONTAINER}" \
  --network "${NETWORK}" \
  -p "${PORT}:8080" \
  -e TFR_DATABASE_HOST="${PG_CONTAINER}" \
  -e TFR_DATABASE_PORT=5432 \
  -e TFR_DATABASE_NAME=terraform_registry \
  -e TFR_DATABASE_USER=registry \
  -e TFR_DATABASE_PASSWORD=registry \
  -e TFR_DATABASE_SSL_MODE=disable \
  -e TFR_SERVER_HOST=0.0.0.0 \
  -e TFR_SERVER_PORT=8080 \
  -e TFR_JWT_SECRET=fetch-spec-only-not-for-production-not-for-production \
  -e ENCRYPTION_KEY=00000000000000000000000000000000 \
  "${IMAGE}" >/dev/null

for i in $(seq 1 60); do
  if curl -sf "http://localhost:${PORT}/openapi3.json" -o "${OUT}.tmp" 2>/dev/null; then
    mv "${OUT}.tmp" "${OUT}"
    echo "    fetched: $(wc -c <"${OUT}") bytes"
    exit 0
  fi
  sleep 1
done

echo "ERROR: backend container never served /openapi3.json"
echo "--- postgres logs ---"
docker logs "${PG_CONTAINER}" 2>&1 | tail -20 || true
echo "--- backend logs ---"
docker logs "${BE_CONTAINER}" 2>&1 | tail -30 || true
exit 1
