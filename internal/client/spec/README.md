# Generated client types

This package contains Go types generated from the backend's OpenAPI 3 spec, and
the vendored spec itself, which the route-contract test in `internal/client`
also checks every client call against.

## What's here

- **`openapi3.json`** — the backend's spec, copied verbatim from the release
  named in `BACKEND_VERSION` by `fetch-spec.sh`. It is the source of truth for
  everything below, and nothing else writes it: never edit it by hand.
- **`BACKEND_VERSION`** — the backend release tag the spec was copied from
  (currently `v1.1.6`). The spec's own `info.version` is a fixed placeholder, so
  this file is what ties the snapshot to a release.
- **`openapi3.json.sha256`** — the SHA-256 `fetch-spec.sh` recorded for that
  copy. The drift job and the unit tests fail if `openapi3.json` stops matching
  it.
- **`openapi3-patched.json`** — the spec after `preprocess.py` runs. Patches:
  - Lifts operation-level path parameters to path level
    (backend [#359](https://github.com/sethbacon/terraform-registry-backend/issues/359)).
  - Dedupes string enum values that swag emitted twice due to Go type aliases
    (backend [#360](https://github.com/sethbacon/terraform-registry-backend/issues/360)).
  - Declares the `SetupToken` security scheme referenced by setup endpoints
    but absent from `components.securitySchemes`
    (backend [#361](https://github.com/sethbacon/terraform-registry-backend/issues/361)).
- **`models_gen.go`** — the generated Go types, package `spec`. ~2700 lines covering
  every schema in `components.schemas`.
- **`fetch-spec.sh`** — maintainer-only; see
  [Moving to another backend release](#moving-to-another-backend-release).

## Why a parallel package

The existing hand-written types in `internal/client/models.go` are still authoritative
for the running provider. Generated types live in a separate `spec` sub-package so
they can coexist while individual resources are migrated one at a time. Each
follow-up PR will switch one resource (e.g. `User`) from the hand-written struct to
`spec.User` and delete the hand-written variant — see
[issue #34](https://github.com/sethbacon/terraform-provider-registry/issues/34) for
the migration plan. So far only `User` has moved.

## Regenerating the types

```bash
# Requires python3 and oapi-codegen v2.7.0, the version CI installs:
#   go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.0
# (another version rewrites the generated header and fails the drift check).
# No network, docker or backend is needed.
make models-gen
```

The Makefile target:

1. Runs `preprocess.py` on the committed `openapi3.json`.
2. Runs `oapi-codegen` against the patched spec.
3. `gofmt`s the result.

CI (the Generated Models Drift job in `.github/workflows/test.yml`) runs the same
target on every PR and fails on `git diff`, so `openapi3-patched.json` and
`models_gen.go` cannot drift from the committed spec. The job needs nothing but
the checkout.

## Route contract

`internal/client/route_contract_test.go` parses the client with `go/ast`, lowers
every request it makes to a method and a path template, and fails when one is not
an operation in `openapi3.json`. It runs with the unit tests
(`go test ./internal/client/...`) and needs no backend.

When it fails for a call you added or changed:

- the path or the method is wrong: fix the client;
- the backend serves the route, but this spec predates it: move the spec to a
  release that documents it (below);
- the route is deliberately missing from the spec: add it to `knownRouteGaps` in
  the test, with the reason. Entries that stop being needed fail the test until
  they are removed.

## Moving to another backend release

Maintainers only; CI never runs this. You need a local clone of
terraform-registry-backend with its tags fetched.

```bash
internal/client/spec/fetch-spec.sh ../terraform-registry-backend v1.1.6
make models-gen
go test ./internal/client/...
```

`fetch-spec.sh` reads `backend/docs/openapi3.json` at the tag straight from the
clone's git objects (the backend embeds that exact file and serves it at
`/openapi3.json`) and rewrites `openapi3.json`, `BACKEND_VERSION` and
`openapi3.json.sha256`. Review the spec diff, since everything in `openapi3.json`
is published with this repository, then commit those three files together with
the regenerated `openapi3-patched.json` and `models_gen.go` in one PR.

## Why we still have a preprocessor

Each of the three patches has a matching backend issue. When all three land
upstream, `preprocess.py` becomes a no-op and can be deleted. Until then, the
preprocessor keeps the local generation green without blocking on upstream merges.
