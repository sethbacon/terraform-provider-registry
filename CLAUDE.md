# CLAUDE.md — Terraform Provider for Private Terraform Registry

## Development Workflow

All changes follow this workflow. Do not deviate from it.

### Branches

- `main` — production-ready, tagged releases only. **Must always exist — never delete.**
- `development` — integration branch; all feature/fix branches merge here first. **Must always exist — never delete.**
- Feature/fix branches are created from `development`, never from `main`. Delete them from remote after their PR is merged; clean up locally with `git branch -d`.

```bash
# After a feature/fix PR is merged:
git push origin --delete fix/short-description   # remove remote branch
git branch -d fix/short-description              # remove local branch
git remote prune origin                          # prune stale remote-tracking refs
```

### Step-by-step

1. **Open a GitHub issue** describing the bug or feature before writing any code.

2. **Create a branch from `development`**:

   ```bash
   git fetch origin
   git checkout -b fix/short-description origin/development
   # or: feature/short-description
   ```

3. **Implement the change.**

4. **Before committing — run the full local quality gate** (CI will reject anything that fails these):

   ```bash
   # Build
   go build ./...

   # Format & vet
   go fmt ./...
   go vet ./...

   # Lint (requires golangci-lint v2)
   golangci-lint run

   # Unit tests
   go test -v -count=1 ./internal/client/...

   # Acceptance tests (requires running backend — see README)
   TF_ACC=1 TF_REGISTRY_ENDPOINT=http://localhost:8081 go test -v ./internal/provider/...
   ```

   Do not push until all of the above pass locally.

5. **Commit — no co-author attribution**:

   ```bash
   git add <specific files>
   git commit -m "fix: short description of what was fixed

   Closes #<issue-number>"
   ```

6. **Rebase onto `development` before pushing** to minimise merge conflicts with sibling branches:

   ```bash
   git fetch origin
   git rebase origin/development
   ```

7. **Push to origin**:

   ```bash
   git push -u origin fix/short-description
   ```

8. **Open a PR from the feature branch → `development`**:

   Include a `## Changelog` section in the PR body with the entry that should appear in `CHANGELOG.md` for this change. **Do not edit `CHANGELOG.md` in the branch** — changelog entries are collected from merged PR bodies at release time.

   ```bash
   gh pr create --base development --title "fix: short description" --body "$(cat <<'EOF'
   Closes #<issue>

   ## Changelog
   - fix: short description of what was fixed
   EOF
   )"
   ```

   - Squash-merge into `development` when approved.

9. **Open a PR from `development` → `main`** when the integration branch is ready to ship:

   ```bash
   gh pr create --base main --title "chore: release vX.Y.Z" --body "..."
   ```

### Parallel agents — coordination rules

When multiple agents run concurrently, follow these rules to avoid conflicts:

- **Never assign two agents to work on the same files at the same time.** If their scopes overlap, serialise them.
- **Do not edit `CHANGELOG.md` in any branch.** Changelog entries live in PR bodies only (see step 8 above). This eliminates the most common parallel-agent conflict.
- **Each agent rebases on `origin/development` immediately before pushing** (step 6 above). After any sibling PR is merged, remaining open branches must rebase again before their own merge.

### Releasing a version

When a release is called for:

1. Collect the `## Changelog` sections from all PR bodies merged since the last release.

2. Update `CHANGELOG.md` on `development` — promote `[Unreleased]` to the new version with today's date and paste the collected entries:

   ```markdown
   ## [X.Y.Z] - YYYY-MM-DD
   ### Fixed
   - fix: ...
   ### Added
   - feat: ...
   ```

3. Commit directly on `development` and push (**no tag yet**):

   ```bash
   git commit -m "chore: release vX.Y.Z"
   git push origin development
   ```

4. Merge `development` → `main` via PR.

5. **After the PR is merged**, tag the commit that landed on `main` and push the tag:

   ```bash
   git fetch origin
   git tag vX.Y.Z origin/main
   git push origin vX.Y.Z
   ```

   > **Why tag after the merge?** The release PR produces a new merge commit SHA on `main`.
   > Tagging on `development` before the merge leaves the tag pointing at the wrong commit —
   > it will never appear in `main`'s history as a tagged release.

6. **Verify the release workflow** fired within ~60 seconds:

   ```bash
   gh run list --workflow=release.yml --limit=3
   ```

   The workflow builds multi-platform provider binaries, creates a GitHub Release, and (if configured) publishes to the Terraform Registry.

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
│   ├── docker-compose.test.yml   # Test stack (backend + postgres)
│   └── seed-dev-admin.sql        # Dev admin user seed
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
docker compose -f deployments/docker-compose.test.yml up -d
TF_ACC=1 TF_REGISTRY_ENDPOINT=http://localhost:8081 go test -v ./internal/provider/...

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
- Acceptance tests require a live backend; the `deployments/docker-compose.test.yml` stack provides one.
- `DEV_MODE=true` on the backend enables `POST /api/v1/dev/login` for test token fetching.
- If `TF_REGISTRY_TOKEN` is unset, `TestMain` in `provider_test.go` fetches a dev token automatically.
- Do not edit `CHANGELOG.md` in feature branches — see releasing workflow above.
