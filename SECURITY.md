# Security Policy

## Supported Versions

| Version | Supported |
| ------- | --------- |
| Latest  | Yes       |
| Older   | No        |

Only the latest published release receives security fixes. If you are running an older version,
please upgrade before reporting.

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Report vulnerabilities privately via [GitHub Security Advisories](https://github.com/sethbacon/terraform-provider-registry/security/advisories/new).

Include:

- A description of the vulnerability and its potential impact.
- Steps to reproduce or a proof-of-concept (where possible).
- The version(s) affected.

You will receive an acknowledgement within **5 business days** and a resolution timeline within
**14 business days** of triage.

## Supply-Chain Verification

Every release binary is signed with GPG and its checksum file is attached to the GitHub Release.
Build provenance is attested via [GitHub Artifact Attestations](https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds).

Verify a downloaded binary:

```bash
# Verify build provenance
gh attestation verify <binary-file> --repo sethbacon/terraform-provider-registry

# Verify GPG signature on the checksum file
gpg --verify terraform-provider-registry_<version>_SHA256SUMS.sig \
              terraform-provider-registry_<version>_SHA256SUMS

# Verify binary checksum
sha256sum --check --ignore-missing terraform-provider-registry_<version>_SHA256SUMS
```

## Scope

This policy covers the Terraform provider code and its release artifacts. Vulnerabilities in
the self-hosted [Terraform Registry Backend](https://github.com/sethbacon/terraform-registry-backend)
should be reported to that repository's security policy.

## Shared CI workflows

Part of this repository's CI is **defined in another repository** — [`4cloudguru/shared-workflows`](https://github.com/4cloudguru/shared-workflows) — and called from `.github/workflows/`. That is a real supply-chain relationship, and it is recorded here so an audit of this repository does not stop at this repository's own tree.

**What runs, and where it is pinned.** Each caller in `.github/workflows/` names the shared workflow on its `uses:` line, pinned to a full 40-hex commit SHA with a trailing comment naming the release that SHA is. The tag is a label; the SHA is what runs. An unlabelled SHA is rejected by the workflow-hardening gate, because a bare 40-hex ref cannot be reviewed or updated deliberately.

**Why the pins have to agree across repositories.** A shared definition drifts differently from a duplicated file: every repository looks like it is using "the shared one" while sitting on different commits, which is *harder* to see than divergent files, not easier. A signature in `security-orchestration` (`shared-workflow-pin-parity`) reports **disagreement** between callers of the same shared workflow — it reports disagreement rather than staleness, because a repository deliberately held back is a decision while N repositories disagreeing without anyone deciding is drift.

**What the shared repository is itself protected by.** Its `main` requires its own zizmor and actionlint checks with `enforce_admins` enabled, restricts which third-party actions may run to an explicit allowlist, issues a read-only default `GITHUB_TOKEN`, and runs the workflow-hardening gate against itself.

**What this repository still controls.** Triggers, concurrency, and the secrets it passes. Secrets are passed **by name** — never `secrets: inherit`, which would forward every secret in this repository to a workflow owned by someone else. Any `vars.*` a shared workflow reads resolve against **this** repository, so credentials and their installation scope do not move.
