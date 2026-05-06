# Terraform Provider for Private Terraform Registry

The **registry** Terraform provider manages all resources in a self-hosted
[Terraform Registry Backend](https://github.com/sethbacon/terraform-registry-backend)
instance — users, organizations, modules, providers, mirrors, SCM integrations,
storage backends, policies, and more.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25 (only required to build the provider from source)
- A running [Terraform Registry Backend](https://github.com/sethbacon/terraform-registry-backend) >= v1.0.0

## Using the Provider

Add the provider to your Terraform configuration:

```terraform
terraform {
  required_providers {
    registry = {
      source  = "sethbacon/registry"
      version = "~> 0.3"
    }
  }
}

provider "registry" {
  endpoint = "https://registry.example.com"
  token    = var.registry_token

  # Optional tuning knobs:
  # insecure      = false  # disable TLS cert verification (dev only)
  # timeout       = 30     # HTTP request timeout in seconds
  # max_retries   = 3      # retries for 429 / 5xx responses
  # version_check = true   # warn if backend is older than the minimum supported version
}

variable "registry_token" {
  description = "API key or JWT token for the registry."
  type        = string
  sensitive   = true
}
```

The `token` and `endpoint` can also be supplied via environment variables:

```shell
export TF_REGISTRY_ENDPOINT="https://registry.example.com"
export TF_REGISTRY_TOKEN="tfr_..."
```

### Provider configuration reference

| Argument | Type | Description |
| --- | --- | --- |
| `endpoint` | string | Base URL of the registry backend. Falls back to `TF_REGISTRY_ENDPOINT`. |
| `token` | string (sensitive) | API key or JWT bearer token. Falls back to `TF_REGISTRY_TOKEN`. |
| `insecure` | bool | Disable TLS certificate verification. Development only. Default `false`. |
| `timeout` | int | HTTP request timeout in seconds. Default `30`. |
| `max_retries` | int | Retries for `429` and `5xx` responses. Default `3`. |
| `version_check` | bool | Probe the backend's `GET /version` and warn if it is older than the minimum supported. Default `true`. |

## Documentation

Full provider documentation is available on the
[Terraform Registry](https://registry.terraform.io/providers/sethbacon/registry/latest/docs).

### Resources

#### Identity and access

| Resource | Description |
| --- | --- |
| `registry_user` | Registry user account |
| `registry_organization` | Organization (namespace) |
| `registry_organization_member` | User membership and role within an organization |
| `registry_api_key` | Scoped API key for CI/CD and automation |
| `registry_role_template` | Custom RBAC role template |
| `registry_oidc_group_mapping` | Maps an external IdP group to a registry role |

#### Modules and providers

| Resource | Description |
| --- | --- |
| `registry_module` | Module record |
| `registry_module_deprecation` | Deprecation marker for an entire module |
| `registry_module_version_deprecation` | Deprecation marker for a single module version |
| `registry_module_reanalyze` | Trigger re-scanning / re-analysis of a module version |
| `registry_module_scm_link` | SCM-to-module link for automatic version publishing |
| `registry_provider_record` | Provider record |
| `registry_provider_version_deprecation` | Deprecation marker for a single provider version |
| `registry_scm_provider` | SCM integration (GitHub, GitLab, Azure DevOps, Bitbucket) |

#### Mirroring, storage, and policy

| Resource | Description |
| --- | --- |
| `registry_mirror` | Provider network mirror configuration |
| `registry_terraform_mirror` | Terraform/OpenTofu binary mirror configuration |
| `registry_storage_config` | Storage backend configuration (local, S3, Azure Blob, GCS) |
| `registry_storage_migration` | Migrate existing artifacts between storage backends |
| `registry_policy` | Mirror approval policy |
| `registry_approval_request` | Mirror approval request |

### Data Sources

#### Identity and access (data sources)

| Data Source | Description |
| --- | --- |
| `registry_users` | List users |
| `registry_organizations` | List organizations |
| `registry_api_keys` | List API keys |
| `registry_role_templates` | List role templates |
| `registry_identity_group_mappings` | List IdP group → role mappings |
| `registry_oidc_config` | Current OIDC SSO configuration |
| `registry_mtls_config` | Current mTLS configuration |

#### Modules, providers, and SCM (data sources)

| Data Source | Description |
| --- | --- |
| `registry_modules` | List module records |
| `registry_providers` | List provider records |
| `registry_scm_providers` | List SCM integrations |
| `registry_advisories` | Security advisories applicable to registry artifacts |
| `registry_scan` | Result of a single scan run |
| `registry_module_scan` | Latest scan result for a module version |
| `registry_scanning_config` | Current scanning configuration |
| `registry_scanning_stats` | Aggregated scanning statistics |

#### Mirroring, policy, and observability (data sources)

| Data Source | Description |
| --- | --- |
| `registry_mirrors` | List provider mirrors |
| `registry_terraform_mirrors` | List Terraform/OpenTofu binary mirrors |
| `registry_terraform_mirror_history` | History of binary mirror sync operations |
| `registry_terraform_mirror_version` | Single Terraform/OpenTofu binary mirror version |
| `registry_terraform_mirror_versions` | Versions available in a Terraform/OpenTofu binary mirror |
| `registry_policy_engine_config` | Current policy engine configuration |
| `registry_policy_evaluation` | Result of a policy evaluation |
| `registry_audit_log` | A single audit log entry by ID |
| `registry_audit_logs` | Query audit log entries |
| `registry_stats` | Registry dashboard statistics |

## Example Usage

The following example bootstraps a complete organization with a module, provider,
API key, and an SCM integration for automatic publishing:

```terraform
resource "registry_organization" "example" {
  name         = "my-org"
  display_name = "My Organization"
}

resource "registry_user" "alice" {
  email = "alice@example.com"
  name  = "Alice"
}

resource "registry_organization_member" "alice" {
  organization_id = registry_organization.example.id
  user_id         = registry_user.alice.id
}

resource "registry_api_key" "ci" {
  organization_id = registry_organization.example.id
  name            = "ci-pipeline"
  scopes          = ["modules:read", "modules:write", "providers:read", "providers:write"]
}

resource "registry_module" "vpc" {
  organization_id = registry_organization.example.id
  namespace       = registry_organization.example.name
  name            = "vpc"
  system          = "aws"
  description     = "AWS VPC module"
  source          = "https://github.com/my-org/terraform-aws-vpc"
}

resource "registry_scm_provider" "github" {
  name = "github-main"
  type = "github"
}

resource "registry_module_scm_link" "vpc" {
  module_id       = registry_module.vpc.id
  scm_provider_id = registry_scm_provider.github.id
  owner           = "my-org"
  repo            = "terraform-aws-vpc"
  branch          = "main"
  tag_pattern     = "v*"
}

output "ci_api_key" {
  value     = registry_api_key.ci.key
  sensitive = true
}
```

## Developing the Provider

### Building

```shell
git clone https://github.com/sethbacon/terraform-provider-registry
cd terraform-provider-registry
go build ./...
```

### Installing locally

```shell
make install
```

This installs the provider binary into the local Terraform plugin directory
(`~/.terraform.d/plugins/registry.terraform.io/terraform-registry/registry/<version>/<os_arch>/`)
so it can be used by a local `terraform` configuration that references
`registry.terraform.io/terraform-registry/registry`.

The `terraform-registry` namespace is a local-only convention used by the
`make install` workflow for development. Released versions are published to
the public Terraform Registry under `sethbacon/registry` (the source path
shown in the `required_providers` block above).

### Running tests

Unit tests (no backend required):

```shell
make test
```

Acceptance tests require a running registry backend. The easiest way is to use
the included test stack:

```shell
# Start the backend (pulls a pinned ghcr.io/sethbacon/terraform-registry-backend tag)
docker compose -f deployments/docker-compose.test.yml up -d

# Seed the dev admin user (run once per fresh database)
docker compose -f deployments/docker-compose.test.yml exec postgres \
  psql -U registry -d terraform_registry < deployments/seed-dev-admin.sql

# Run acceptance tests
make testacc
```

The `TF_REGISTRY_ENDPOINT` defaults to `http://localhost:8081`.
If `TF_REGISTRY_TOKEN` is not set, the test suite fetches a token automatically
via the backend's dev login endpoint (requires `DEV_MODE=true`, which the test
stack enables by default).

### Generating documentation

Provider documentation is generated from schema descriptions and template files
using [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs):

```shell
make docs
```

## Contributing

This project uses a single-branch model: all `fix/*` and `feature/*` branches
are squash-merged directly into `main` via pull request.

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full workflow — branching
conventions, local quality gate, commit style, PR checklist — and
[RELEASING.md](RELEASING.md) for the release process.

## License

This provider is licensed under the [Apache License 2.0](LICENSE).
