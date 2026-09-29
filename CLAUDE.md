# CLAUDE.md — Terraform Provider for Private Terraform Registry

## Development Workflow

### Branch Model

Single-branch: `main` only. Push directly with `--admin` bypass. No PRs required.

No co-author attribution in commits.

### Before pushing — quality gate

```bash
go build ./...
go fmt ./...
go vet ./...
golangci-lint run
go test -v -count=1 ./internal/client/...
```

Acceptance tests (optional, requires running backend):
```bash
TF_ACC=1 TF_REGISTRY_ENDPOINT=http://localhost:8081 go test -v ./internal/provider/...
```

### Commit style

```bash
git commit -m "fix: short description

Closes #<issue-number>"
```

### Releasing a version

1. Tag the commit on `main` and push:
   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
2. The `release.yml` workflow builds multi-platform provider binaries via GoReleaser and creates a GitHub Release.

---

## Project Overview

A Terraform provider for managing all resources in a self-hosted
[Terraform Registry Backend](https://github.com/sethbacon/terraform-registry-backend):
users, organizations, modules, providers, mirrors, SCM integrations, storage backends, policies, and more.

---

## Repository Structure

```txt
terraform-provider-registry/
├── internal/
│   ├── client/       # HTTP client for the registry backend API
│   └── provider/     # Terraform resource and data source implementations
├── deployments/
│   └── docker-compose.test.yml   # Test stack (backend + postgres); TFR_ACC_BACKEND_IMAGE names the backend image
├── .github/workflows/
│   ├── test.yml      # CI: build, lint, unit tests, acceptance tests
│   └── release.yml   # GoReleaser: triggered by vX.Y.Z tag push
├── .golangci.yml     # golangci-lint v2 configuration
├── go.mod / go.sum
├── main.go
└── CHANGELOG.md
```

---

## Common Commands

```bash
# Build
go build ./...

# Unit tests (no backend required)
go test -v -count=1 ./internal/client/...

# Acceptance tests (requires backend — see README)
TFR_ACC_BACKEND_IMAGE=<image> docker compose -f deployments/docker-compose.test.yml up -d
TF_ACC=1 TF_REGISTRY_ENDPOINT=http://localhost:8081 \
  TFR_ACC_SETUP_TOKEN_FILE="$(pwd)/deployments/.acc-run/setup-token" \
  go test -v ./internal/provider/...

# Lint
golangci-lint run

# Install locally
make install

# Generate documentation
make docs
```

---

## Tech Stack

| Concern       | Technology                                         |
| ------------- | -------------------------------------------------- |
| Language      | Go 1.25+                                           |
| Framework     | terraform-plugin-framework (hashicorp)             |
| HTTP client   | net/http with retry + backoff                      |
| Lint          | golangci-lint v2 (.golangci.yml)                   |
| Release       | GoReleaser (triggered by vX.Y.Z tag on `main`)     |
| Docs          | terraform-plugin-docs                              |

---

## Development Notes

- The provider is published as `sethbacon/registry` on the Terraform Registry.
- Acceptance tests require a live backend, any release — `TFR_ACC_BACKEND_IMAGE` and
  `deployments/docker-compose.test.yml` provide one for local/CI use, or point
  `TF_REGISTRY_ENDPOINT` at any other registry.
- If `TF_REGISTRY_TOKEN` is unset, `TestMain` in `provider_test.go` bootstraps a
  platform-admin account through the backend's setup API (using the one-time setup
  token at `TFR_ACC_SETUP_TOKEN_FILE`) and signs in via `DEV_MODE`'s dev-login
  endpoint. No database writes.
