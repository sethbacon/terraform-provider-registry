# Releasing

This document describes the end-to-end release process for `terraform-provider-registry`.

## Overview

Releases are **manual** — a maintainer collects changelog entries from merged PRs, bumps the
version in `CHANGELOG.md` via a short-lived branch PR to `main`, then pushes a semver tag.
GoReleaser builds and publishes binaries automatically on tag push.

## Prerequisites

- Write access to the repository.
- GPG key loaded (`GPG_PRIVATE_KEY` and `PASSPHRASE` secrets configured in GitHub Actions).
- `gh` CLI authenticated.

## Step-by-step

### 1. Collect changelog entries

Scan merged PR bodies since the last release for `## Changelog` sections. Each merged PR that
touches provider behaviour should have contributed one or more bullet points.

### 2. Open a changelog PR to `main`

Create a short-lived branch, update `CHANGELOG.md`, and open a PR:

```bash
git fetch origin
git checkout -b chore/release-vX.Y.Z origin/main
```

Add a new section at the top of `CHANGELOG.md` (below `## [Unreleased]`):

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- ...

### Fixed
- ...

### Changed
- ...
```

```bash
git add CHANGELOG.md
git commit -m "chore: release vX.Y.Z"
git push -u origin chore/release-vX.Y.Z
gh pr create \
  --title "chore: release vX.Y.Z" \
  --body "Release vX.Y.Z — see CHANGELOG.md for details." \
  --base main \
  --head chore/release-vX.Y.Z
```

### 3. Merge the release PR

CI must pass. Squash-merge:

```bash
gh pr merge <PR-number> --squash --admin --delete-branch
```

### 4. Tag the release

```bash
git fetch origin main
git tag vX.Y.Z origin/main
git push origin vX.Y.Z
```

The `release.yml` workflow fires automatically on the tag push.

### 5. Verify the release

```bash
gh run list --workflow=release.yml --limit 3
gh release view vX.Y.Z
```

Check that:

- GoReleaser completed successfully.
- Binaries and checksums are attached to the GitHub Release.
- Build provenance attestation is present:

  ```bash
  gh attestation verify <binary> --repo sethbacon/terraform-provider-registry
  ```

## Hotfix releases

For urgent fixes on an already-shipped version:

1. Branch from the release tag: `git checkout -b fix/short-description vX.Y.Z`
2. Apply the fix, update `CHANGELOG.md` with a new patch entry.
3. Push the branch and open a PR targeting `main`.
4. After merge, tag and push `vX.Y.Z+1`.

## Manual fallback (GoReleaser failure)

If GoReleaser fails after the tag is pushed:

1. Fix the underlying issue (workflow, config, secret).
2. Delete the tag: `git push origin --delete vX.Y.Z`
3. Re-push the tag: `git tag vX.Y.Z origin/main && git push origin vX.Y.Z`

Do **not** edit published release assets manually — re-run the workflow instead.

## Supply-chain verification

Every release binary is GPG-signed and has GitHub Artifact Attestation build provenance.
See [SECURITY.md](SECURITY.md) for verification commands.

## Terraform Registry publication

GoReleaser automatically attaches the provider manifest
(`terraform-provider-registry_vX.Y.Z_manifest.json`) and all platform archives in the format
expected by the Terraform CLI.
