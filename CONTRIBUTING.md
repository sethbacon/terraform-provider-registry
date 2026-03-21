# Contributing

Thank you for your interest in contributing to the Terraform Registry Provider!

## Branching Model

This project uses a two-branch strategy:

- **`main`** — production-ready code only. Every commit on `main` corresponds to a tagged release.
  Do not push directly to `main`.
- **`development`** — the integration branch. All feature and fix branches are merged here first.
  CI must pass before merging.

Feature and fix branches are created from `development` and merged back into `development` via
pull request. When the integration branch is stable and ready to ship, a release PR is opened
from `development` → `main`.

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) >= 1.25
- [golangci-lint](https://golangci-lint.run/welcome/install/) v2.x
- [Docker](https://docs.docker.com/get-docker/) (for acceptance tests)
- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0 (for local testing)

### Setup

```bash
git clone https://github.com/sethbacon/terraform-provider-registry
cd terraform-provider-registry
go mod download
```

## Making a Change

### 1. Open an issue

Before writing code, open a GitHub issue describing the bug or feature. Reference this issue
in your commits and PR body.

### 2. Create a branch from `development`

```bash
git fetch origin
git checkout -b fix/short-description origin/development
# or: feature/short-description
```

Use a short, lowercase, hyphenated description. Prefix with `fix/` for bug fixes or
`feature/` for new functionality.

### 3. Implement and test

Run the full quality gate locally before pushing:

```bash
# Build
go build ./...

# Format and vet
go fmt ./...
go vet ./...

# Lint
golangci-lint run

# Unit tests (no backend required)
go test -v -count=1 ./internal/client/...
```

For changes that affect provider behaviour, run the acceptance tests too:

```bash
# Start the test backend
docker compose -f deployments/docker-compose.test.yml up -d

# Seed the dev admin user (once per fresh database)
docker compose -f deployments/docker-compose.test.yml exec -T postgres \
  psql -U registry -d terraform_registry < deployments/seed-dev-admin.sql

# Run acceptance tests
TF_ACC=1 TF_REGISTRY_ENDPOINT=http://localhost:8081 go test -v ./internal/provider/...
```

Do not push until all checks pass locally.

### 4. Commit

Write clear, imperative commit messages. No co-author attribution lines.

```bash
git add <specific files>
git commit -m "fix: short description

Closes #<issue-number>"
```

### 5. Rebase before pushing

```bash
git fetch origin
git rebase origin/development
```

### 6. Push and open a pull request

```bash
git push -u origin fix/short-description
```

Open a PR targeting **`development`** (not `main`). Include:

- A description of the change and why it is needed.
- A reference to the issue (`Closes #N`).
- A `## Changelog` section with the entry for `CHANGELOG.md`:

```markdown
## Changelog
- fix: short description of what was fixed
```

Do **not** edit `CHANGELOG.md` in your branch. Changelog entries are collected from merged PR
bodies at release time.

### 7. Review and merge

- Ensure CI passes (Build, Lint, Unit Tests, Acceptance Tests).
- Address review feedback.
- A maintainer will squash-merge your PR into `development`.

### 8. Clean up

After your PR is merged:

```bash
git push origin --delete fix/short-description   # remove remote branch
git branch -d fix/short-description              # remove local branch
git remote prune origin                          # prune stale remote-tracking refs
```

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Keep golangci-lint (`golangci-lint run`) clean — no new warnings.
- Separate third-party imports from internal imports with a blank line:

  ```go
  import (
      "context"
      "fmt"

      "github.com/hashicorp/terraform-plugin-framework/resource"

      "github.com/terraform-registry/terraform-provider-registry/internal/client"
  )
  ```

- Wrap `defer resp.Body.Close()` to satisfy errcheck:

  ```go
  defer func() { _ = resp.Body.Close() }()
  ```

## Running the Full Test Suite

Unit tests only:

```bash
make test
```

Acceptance tests (requires Docker):

```bash
make testacc
```

## Releasing

Releases are managed by maintainers. If you believe a release is needed, open an issue or
mention it in a PR discussion.

The release process:

1. Collect changelog entries from merged PR bodies since the last release.
2. Update `CHANGELOG.md` on `development`.
3. Open a PR from `development` → `main`.
4. After the PR is merged, tag the commit on `main` with `vX.Y.Z`.
5. The `release.yml` workflow runs GoReleaser automatically.

## Questions

Open a [GitHub issue](https://github.com/sethbacon/terraform-provider-registry/issues) for
bug reports, feature requests, or questions.
