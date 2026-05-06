<!-- markdownlint-disable MD024 -->

# Changelog

All notable changes to this provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

---

## [0.3.2] - 2026-05-06

### Fixed

- `registry_module_version_deprecation` no longer churns state on every refresh.
  `client.GetModuleVersion` was hitting `GET /api/v1/modules/{ns}/{name}/{sys}/{version}`
  — a route the backend does not expose — so the Read implementation always
  saw a 404, removed the resource from state, and triggered recreation on the
  next plan. The client now derives version-level state from the existing
  `GET /api/v1/modules/{ns}/{name}/{sys}` endpoint, the same approach the
  frontend uses. Works against any backend version (#74, #75)

### Changed

- Bumped `GNUmakefile` `VERSION` to `0.3.2` so `make install` drops the
  binary into a path matching the published release.

---

## [0.3.1] - 2026-05-06

### Documentation

- README aligned with current provider state: corrected version constraint
  (`~> 0.3`), documented the `insecure` / `timeout` / `max_retries` /
  `version_check` provider arguments, added 6 missing resources and 14
  missing data sources to the resource tables, clarified that the
  `terraform-registry/registry` source path is a local-only convention
  used by `make install`, and replaced the stale two-branch model note
  with the current single-branch flow (#72)
- Bumped `GNUmakefile` `VERSION` to match this release so `make install`
  drops the binary into a path matching the published provider version (#72)

### Changed

- Acceptance-test CI is now sharded across 4 parallel jobs, cutting wall
  time from ~40 min to ~12 min per PR. An aggregator job preserves the
  `Acceptance Tests` required-check name so branch protection is unchanged (#71)

### Fixed

- Bumped `go.opentelemetry.io/otel` and `go.opentelemetry.io/otel/sdk` to
  `v1.43.0`, resolving four transitive OSV findings (GHSA-mh2q-q3fh-2475,
  GHSA-9h8m-3fm2-qjrq, GHSA-hfvc-g4fc-pqhx, GO-2026-4394) (#70)
- Added `.gitattributes` to enforce LF line endings on Windows checkouts
  so `gofmt` no longer reports false-positive lint failures; corrected
  `const` alignment in `resource_storage_migration.go` (#68)
- Applied `gofmt` to three files; pinned the test-stack backend image to
  the correct registry tag (`1.0.5`, no `v` prefix) (#67)

### CI / supply chain

- Corrected pinned action SHAs across all workflows so jobs run against
  the intended action versions (#65)
- Aligned the repo's CI hardening posture with the backend: pinned
  action SHAs everywhere, weekly OSV-Scanner + CodeQL job, dependency
  review on PRs, Conventional PR-title check, dependabot for `gomod`
  and `github-actions`, SLSA build-provenance attestation on releases (#63)

---

## [0.3.0] - 2026-05-05

### Added

- `registry_module_deprecation` — marks a module as deprecated; `message` and optional `successor_module_id`; destroy removes the deprecation (#54)
- `registry_module_version_deprecation` — marks a specific module version as deprecated; supports `replacement_source` for Terraform CLI ≥1.10 upgrade guidance (#54)
- `registry_provider_version_deprecation` — marks a specific provider version as deprecated (#54)
- `registry_oidc_group_mapping` — manages a single OIDC group → organization role mapping (read-modify-write against full-replace backend endpoint) (#55)
- `data.registry_oidc_config` — reads OIDC issuer, client_id, scopes, groups_claim, username_claim (#55)
- `registry_storage_migration` — triggers a one-shot storage backend migration with configurable `timeout_minutes` (default 60) and progress counters; destroy cancels if still running (#58)
- `registry_module_reanalyze` — trigger-pattern resource that re-runs terraform-docs / scanning / SCM verification on a module version; recreates when `triggers` map changes (#61)
- `data.registry_terraform_mirror_versions` — list all mirrored Terraform/OpenTofu versions for a mirror config (#56)
- `data.registry_terraform_mirror_version` — single version detail including available platforms (#56)
- `data.registry_terraform_mirror_history` — sync history entries for a mirror (#56)
- `data.registry_policy_engine_config` — reads rego bundle URL, ETag, status, and last_loaded_at (#57)
- `data.registry_policy_evaluation` — evaluates `input_json` against the rego bundle at plan time; returns `allowed`, `reason`, `result_json` (#57)
- `data.registry_audit_log` — fetches a single audit log entry by UUID with `metadata_json` (#59)
- `data.registry_identity_group_mappings` — reads SAML/LDAP group → role mappings from backend runtime config (#59)
- `data.registry_mtls_config` — reads mTLS enabled state, client CA CN, and server cert subject (#59)
- `data.registry_advisories` — lists CVE advisories with `active_only` and `severities` filters (#60)
- `data.registry_scanning_config` — scanner tool configuration (enabled, binary_path, detected_version, enabled_tools) (#60)
- `data.registry_scanning_stats` — aggregate scan statistics by severity (#60)
- `data.registry_scan` — fetches a single security scan by UUID including findings and execution_log (#60)
- `data.registry_module_scan` — fetches the most recent scan for a specific module version (#60)
- Provider now probes `GET /version` on configure and emits a warning diagnostic if the backend is older than the minimum supported version (`1.0.0`); opt out with `version_check = false` (#53)

### Changed

- `deployments/docker-compose.test.yml` backend image pinned from `latest` to `v1.0.0`; pin policy documented in CONTRIBUTING (#52)

---

## [0.2.1] - 2026-05-04

### Added

- `registry_organization` and `data.registry_organizations` support `idp_type` and `idp_name` for organization-IdP binding (#12)
- `registry_module` and `data.registry_modules` expose computed deprecation state: `deprecated`, `deprecated_at`, `deprecation_message`, `successor_module_id` (#13)
- `registry_mirror` and `data.registry_mirrors` support `pull_through_enabled` and `pull_through_cache_ttl_hours` (#14)
- `registry_terraform_mirror` and `data.registry_terraform_mirrors` support `custom_gpg_key` (sensitive) and `skip_gpg_verify` (#15)
- `registry_storage_config` exposes four computed credential-set indicators: `azure_account_key_set`, `s3_access_key_id_set`, `s3_secret_access_key_set`, `gcs_credentials_json_set`; S3 IAM-role auth fields (`s3_role_arn`, `s3_role_session_name`, `s3_external_id`, `s3_web_identity_token_file`) are now accepted in the `config` map (#16)
- `registry_policy` supports `organization_id` (org-scoped policies; null = global), `organization_name` (read-only), and audit fields `created_by` / `created_by_name` (#17)

---

## [0.2.0] - 2026-05-01

This release realigns the provider with backend `terraform-registry-backend` v1.0.0
(the backend was at ~v0.5 when v0.1.0 was cut). Several fields were silently
dropping data or sending unintended writes against the v1 backend; those are
fixed here.

### Breaking Changes

- `data.registry_stats` reshaped to match backend `admin.DashboardStats`. The flat `total_*` counters (`total_modules`, `total_providers`, `total_users`, `total_organizations`, `total_mirrors`, `total_api_keys`) were always returning zero against backend ≥ v0.x and have been replaced with nested objects. Migration table:

  | Old (always 0)        | New                                               |
  |-----------------------|---------------------------------------------------|
  | `total_modules`       | `modules.total`                                   |
  | `total_providers`     | `providers.total`                                 |
  | `total_users`         | `users`                                           |
  | `total_organizations` | `organizations`                                   |
  | `total_mirrors`       | `provider_mirrors.total` + `binary_mirrors.total` |
  | `total_api_keys`      | removed (backend does not expose this counter)    |

  New nested attributes also expose `modules.{versions, downloads, by_system}`, full provider breakdown (manual vs. mirrored), mirror health, binary mirror platforms, and a `recent_syncs` ledger.

- `registry_scm_provider` removed the phantom `oauth_status` attribute (the backend does not and never did expose this field — it was always `null`). Use the new `is_active` attribute, plus `organization_id`, `tenant_id`, `client_id`, and `webhook_secret`, to manage SCM integrations.

### Fixed

- `registry_storage_config` no longer shows perpetual diff on the `active` attribute. The backend renamed the response field from `active` to `is_active` before this provider's first release; the JSON tag has been updated to match. (#6)
- `registry_mirror` and `registry_terraform_mirror` updates no longer overwrite `enabled`, `sync_interval_hours`, `gpg_verify`, or `stable_only` with zero values when those attributes are unchanged. The client request structs now use pointer fields so omitted attributes leave the existing value unchanged on the backend. (#8)
- `registry_scm_provider` updates no longer overwrite unrelated fields with zero values (same root cause as #8). (#9)

### Added

- `registry_scm_provider` exposes `is_active`, `organization_id`, `tenant_id`, `client_id`, and `webhook_secret`. The phantom `oauth_status` field is removed (see breaking changes). (#9)
- `registry_module_scm_link` adds `repository_path` (default `/`, for monorepo layouts) and `auto_publish_enabled` (default `false`, to opt into auto-publish on tag push). The read-side response now also surfaces the persisted `module_path`. (#10)
- `registry_approval_request` exposes 8 new computed attributes: `organization_id`, `requested_by`, `requested_by_name`, `reviewer_name`, `reviewed_at`, `expires_at`, `auto_approved`, and `mirror_name`. (#11)

### Tests

- New client unit tests round-trip representative backend response payloads to guard against future drift on storage configs, mirrors, terraform mirrors, SCM providers, module SCM links, approval requests, and dashboard stats.

---

## [0.1.1] - 2026-03-21

### Changed

- Add `subcategory` frontmatter to provider docs index for registry.terraform.io categorisation

---

## [0.1.0] - 2026-03-21

Initial release of the Terraform Registry provider.

### Resources

- `registry_user` — manage registry user accounts
- `registry_organization` — manage organizations (namespaces)
- `registry_organization_member` — manage user membership and roles within an organization
- `registry_api_key` — manage scoped API keys for CI/CD and automation
- `registry_role_template` — manage custom RBAC role templates with fine-grained scopes
- `registry_module` — manage module records; versions uploaded via registry API or SCM webhook
- `registry_provider_record` — manage provider records; binaries uploaded via registry API
- `registry_module_scm_link` — link a module to an SCM repository for automatic version publishing on Git tag push
- `registry_scm_provider` — manage SCM integrations (GitHub, GitLab, Azure DevOps, Bitbucket)
- `registry_mirror` — manage provider network mirror configurations
- `registry_terraform_mirror` — manage Terraform/OpenTofu binary mirror configurations
- `registry_storage_config` — manage storage backend configurations (local, S3, Azure Blob, GCS)
- `registry_policy` — manage mirror approval policies
- `registry_approval_request` — submit mirror approval requests for admin review

### Data Sources

- `registry_users` — list registry users with optional search filter
- `registry_organizations` — list organizations with optional search filter
- `registry_api_keys` — list API keys with optional user filter
- `registry_modules` — list modules with optional namespace and search filters
- `registry_providers` — list provider records with optional namespace and search filters
- `registry_scm_providers` — list all SCM provider integrations
- `registry_mirrors` — list all provider mirror configurations
- `registry_terraform_mirrors` — list all Terraform/OpenTofu binary mirror configurations
- `registry_role_templates` — list all RBAC role templates
- `registry_audit_logs` — query audit log entries with optional resource type, action, and pagination filters
- `registry_stats` — read registry dashboard statistics

### Provider

- `endpoint` — base URL of the registry backend; also via `TF_REGISTRY_ENDPOINT`
- `token` — API key or JWT bearer token; also via `TF_REGISTRY_TOKEN`
- `insecure` — disable TLS verification (development only)
- `timeout` — HTTP request timeout in seconds (default: 30)
- `max_retries` — max retries for 429 / 5xx responses (default: 3)
