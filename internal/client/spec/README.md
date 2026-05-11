# Generated client types

This package contains Go types generated from the backend's OpenAPI 3 spec.

## What's here

- **`openapi3.json`** — the upstream spec, pulled from a pinned backend image
  (currently `ghcr.io/sethbacon/terraform-registry-backend:1.1.6`). This is the
  authoritative source; consumers should treat it as committed-but-machine-managed.
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

## Why a parallel package

The existing hand-written types in `internal/client/models.go` are still authoritative
for the running provider. Generated types live in a separate `spec` sub-package so
they can coexist while individual resources are migrated one at a time. Each
follow-up PR will switch one resource (e.g. `User`) from the hand-written struct to
`spec.User` and delete the hand-written variant — see
[issue #34](https://github.com/sethbacon/terraform-provider-registry/issues/34) for
the migration plan.

The first deliverable (this PR) is the toolchain only. Nothing in the running provider
imports `spec` yet.

## Regenerating

```bash
# Requires docker, python3, oapi-codegen (installed via tools.go).
make models-gen
```

The Makefile target:

1. Starts a one-shot container of the pinned backend image.
2. Pulls `/openapi3.json` and writes it next to `preprocess.py`.
3. Runs `preprocess.py` to patch the spec.
4. Runs `oapi-codegen` against the patched spec.
5. `gofmt`s the result.

CI runs this on every PR and fails on `git diff`, so drift between the pinned
backend's spec and the committed `models_gen.go` is caught at review time.

## Bumping the backend pin

Edit `BACKEND_TAG` in `fetch-spec.sh` and rerun `make models-gen`. Commit the
resulting `openapi3.json`, `openapi3-patched.json`, and `models_gen.go` together
in one PR — these three files are always in lockstep.

## Why we still have a preprocessor

Each of the three patches has a matching backend issue. When all three land
upstream, `preprocess.py` becomes a no-op and can be deleted. Until then, the
preprocessor keeps the local generation green without blocking on upstream merges.
