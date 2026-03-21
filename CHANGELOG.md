<!-- markdownlint-disable MD024 -->

# Changelog

All notable changes to this provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
