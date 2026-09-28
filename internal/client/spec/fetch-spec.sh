#!/usr/bin/env bash
# fetch-spec.sh — vendor a backend release's OpenAPI 3 spec into this package.
#
# MAINTAINER-ONLY. CI never runs this script. The committed openapi3.json is
# the source of truth: the Generated Models Drift job regenerates the Go types
# from it offline (`make models-gen`), and the route-contract test in
# internal/client checks every client call against it. Run this only to move
# the vendored spec to another backend release.
#
# Usage:
#   internal/client/spec/fetch-spec.sh <backend-checkout> <release-tag>
#
#   <backend-checkout>  a local clone of terraform-registry-backend. Only its
#                       git objects are read, so its current branch and working
#                       tree do not matter; run `git fetch --tags` there first.
#   <release-tag>       the backend release to vendor, e.g. v1.1.6.
#
# It writes, next to this script:
#   openapi3.json         backend/docs/openapi3.json exactly as committed at the
#                         tag.
#   BACKEND_VERSION       the tag, so the spec can always be traced back to the
#                         release it came from (the spec's own info.version is
#                         a fixed placeholder and says nothing).
#   openapi3.json.sha256  the file's SHA-256. The route-contract test refuses a
#                         spec that no longer matches it, so openapi3.json
#                         cannot be hand-edited into agreement with the client
#                         and still claim to be that release's spec.
#
# The backend embeds that same file in its binary and serves it at
# /openapi3.json, so this is the document a running server of the release
# returns, obtained without docker, a database or registry credentials. That is
# why the spec is read from git: the public CI of this repository never needs
# the backend at all.
#
# Afterwards run `make models-gen` and `go test ./internal/client/...`, review
# the spec diff (everything in openapi3.json is published with this
# repository), and commit openapi3.json, BACKEND_VERSION, openapi3.json.sha256,
# openapi3-patched.json and models_gen.go together.

set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <backend-checkout> <release-tag>" >&2
  exit 2
fi

BACKEND_DIR="$1"
TAG="$2"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SPEC_PATH="backend/docs/openapi3.json"

if ! git -C "${BACKEND_DIR}" rev-parse --git-dir >/dev/null 2>&1; then
  echo "ERROR: ${BACKEND_DIR} is not a git checkout" >&2
  exit 1
fi

# A tag, not a branch or a bare commit: BACKEND_VERSION must name something
# that resolves to the same bytes for whoever checks it later.
if ! git -C "${BACKEND_DIR}" rev-parse -q --verify "refs/tags/${TAG}^{commit}" >/dev/null; then
  echo "ERROR: ${TAG} is not a tag in ${BACKEND_DIR} (run git fetch --tags there?)" >&2
  exit 1
fi

if ! git -C "${BACKEND_DIR}" cat-file -e "refs/tags/${TAG}:${SPEC_PATH}" 2>/dev/null; then
  echo "ERROR: ${TAG} has no ${SPEC_PATH} (the backend commits it from v1.1.5 on)" >&2
  exit 1
fi

tmp="$(mktemp "${SCRIPT_DIR}/.openapi3.json.XXXXXX")"
trap 'rm -f "${tmp}"' EXIT

# Carriage returns are stripped so the checksum covers the LF-only bytes git
# stores for *.json in this repository (.gitattributes). JSON never needs a raw
# CR, so this cannot change the document.
git -C "${BACKEND_DIR}" show "refs/tags/${TAG}:${SPEC_PATH}" | tr -d '\r' >"${tmp}"
mv "${tmp}" "${SCRIPT_DIR}/openapi3.json"

# Hash from stdin and write the sha256sum line ourselves: sha256sum and
# `shasum -a 256` mark files differently across platforms (" *name" on
# Windows), and the committed file must not churn with the maintainer's OS.
if command -v sha256sum >/dev/null 2>&1; then
  sum="$(sha256sum <"${SCRIPT_DIR}/openapi3.json" | cut -d' ' -f1)"
else
  sum="$(shasum -a 256 <"${SCRIPT_DIR}/openapi3.json" | cut -d' ' -f1)"
fi
printf '%s  openapi3.json\n' "${sum}" >"${SCRIPT_DIR}/openapi3.json.sha256"
printf '%s\n' "${TAG}" >"${SCRIPT_DIR}/BACKEND_VERSION"

echo "    backend: ${TAG} ($(git -C "${BACKEND_DIR}" rev-parse --short "refs/tags/${TAG}^{commit}"))"
echo "    wrote:   openapi3.json ($(wc -c <"${SCRIPT_DIR}/openapi3.json") bytes), BACKEND_VERSION, openapi3.json.sha256"
echo "    next:    make models-gen && go test ./internal/client/..."
