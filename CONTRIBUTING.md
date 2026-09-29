# Contributing

Thank you for your interest in contributing to the Terraform Registry Provider!

## Branching Model

This project uses a single-branch model — all development happens on feature and fix branches
that are squash-merged directly into `main` via pull request.

- **`main`** — the single integration and release branch. All PRs target `main`. Do not push
  directly to `main`.
- **`fix/<description>`** — bug fix branches.
- **`feature/<description>`** — new functionality branches.

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

### 2. Create a branch from `main`

```bash
git fetch origin
git checkout -b fix/short-description origin/main
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

# Regenerate docs/ (needs terraform on PATH); CI fails if docs/ changes
make docs
git status --porcelain -- docs/
```

Any change to a schema description, `templates/` or `examples/` must be committed
together with the regenerated `docs/`. A new resource or data source needs an
example under `examples/resources/<name>/resource.tf` (plus `import.sh` if it
supports import) or `examples/data-sources/<name>/data-source.tf`.

For changes that affect provider behaviour, run the acceptance tests too:

The test stack in `deployments/docker-compose.test.yml` does not pin a backend image;
point `TFR_ACC_BACKEND_IMAGE` at whatever backend build you want to test against (a
local build, a release tag, or a private registry image). See "Running tests" in
README.md for the full environment-variable contract, including how to run the
suite against an already-configured registry instead of the test stack.

```bash
# Start the test backend
TFR_ACC_BACKEND_IMAGE=<your-backend-image:tag> docker compose -f deployments/docker-compose.test.yml up -d

# Run acceptance tests. No seeding step: TestMain bootstraps a platform-admin
# account through the backend's own setup API, using the one-time setup token
# the backend writes to deployments/.acc-run/ (bind-mounted in
# docker-compose.test.yml). Nothing is written to the database directly.
TF_ACC=1 TF_REGISTRY_ENDPOINT=http://localhost:8081 \
  TFR_ACC_SETUP_TOKEN_FILE="$(pwd)/deployments/.acc-run/setup-token" \
  go test -v ./internal/provider/...
```

Do not push until all checks pass locally.

### 4. Commit

Write clear, imperative commit messages following [Conventional Commits](https://www.conventionalcommits.org/).
No co-author attribution lines.

```bash
git add <specific files>
git commit -m "fix: short description

Closes #<issue-number>"
```

### 5. Rebase before pushing

```bash
git fetch origin
git rebase origin/main
```

### 6. Push and open a pull request

```bash
git push -u origin fix/short-description
```

Open a PR targeting **`main`**. The PR title must follow Conventional Commits — CI enforces
this via the `Conventional PR Title` check. Include:

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

- Ensure CI passes (Build, Lint, Unit Tests, Acceptance Tests, Conventional PR Title,
  Dependency Review).
- Address review feedback.
- A maintainer will squash-merge your PR into `main`.

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
mention it in a PR discussion. See [RELEASING.md](RELEASING.md) for the full process.

## Questions

Open a [GitHub issue](https://github.com/sethbacon/terraform-provider-registry/issues) for
bug reports, feature requests, or questions.
