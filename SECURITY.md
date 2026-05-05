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
