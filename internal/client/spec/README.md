# Vendored backend spec

This directory holds a verbatim copy of the backend's OpenAPI 3 spec. It no
longer generates Go types: the client's hand-written types in
`internal/client/models.go` are authoritative, and the spec's only consumer
is the route-contract test in `internal/client`.

## What's here

- **`openapi3.json`** — the backend's spec, copied verbatim from the release
  named in `BACKEND_VERSION` by `fetch-spec.sh`. Nothing writes it by hand:
  never edit it directly.
- **`BACKEND_VERSION`** — the backend release tag the spec was copied from
  (currently `v1.1.6`). The spec's own `info.version` is a fixed placeholder,
  so this file is what ties the snapshot to a release.
- **`openapi3.json.sha256`** — the SHA-256 `fetch-spec.sh` recorded for that
  copy. `TestVendoredSpecIsTheRecordedBackendRelease` in
  `internal/client/route_contract_test.go` fails if `openapi3.json` stops
  matching it, so the vendored copy can't be hand-edited into agreement with
  the client and still claim to be that release's spec.
- **`fetch-spec.sh`** — maintainer-only; see
  [Moving to another backend release](#moving-to-another-backend-release).

## Why the spec is vendored at all

`internal/client/route_contract_test.go` parses the client with `go/ast`,
lowers every request it makes to a method and a path template, and fails when
one is not an operation in `openapi3.json`. That catches routes the client
gets wrong before they reach a user: two shipped this way and stayed broken
until users hit them (see the test file's header comment for both). The test
needs a copy of the spec to check against, and vendoring one here means CI
runs the check with nothing but the checkout — no backend, no docker, no
registry login.

The test reads `openapi3.json` directly; nothing patches or preprocesses it,
so there's no separate generation step to keep in sync.

## Route contract

Run it with the rest of the unit tests:

```bash
go test ./internal/client/...
```

When it fails for a call you added or changed:

- the path or the method is wrong: fix the client;
- the backend serves the route, but this spec predates it: move the spec to
  a release that documents it (below);
- the route is deliberately missing from the spec: add it to
  `knownRouteGaps` in the test, with the reason. Entries that stop being
  needed fail the test until they are removed.

## Moving to another backend release

Maintainers only; CI never runs this. You need a local clone of
terraform-registry-backend with its tags fetched.

```bash
internal/client/spec/fetch-spec.sh ../terraform-registry-backend v1.1.6
go test ./internal/client/...
```

`fetch-spec.sh` reads `backend/docs/openapi3.json` at the tag straight from
the clone's git objects (the backend embeds that exact file and serves it at
`/openapi3.json`) and rewrites `openapi3.json`, `BACKEND_VERSION` and
`openapi3.json.sha256`. Review the spec diff, since everything in
`openapi3.json` is published with this repository, then commit those three
files together in one PR.

## History

This directory used to also generate Go types from the spec with
oapi-codegen, kept in a separate `spec` package so resources could migrate to
them one at a time (see issue #34). Only `User` ever moved, and the
route-contract test above already validates that every client call names a
real backend route without needing generated types to do it, so the
generator, its preprocessing step and the generated package were dropped;
`internal/client/models.go` has the hand-written `User` type client code
uses now. The spec itself stays vendored, for the reason above.
