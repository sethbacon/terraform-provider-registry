resource "registry_api_key" "ci" {
  organization_id = registry_organization.example.id
  name            = "ci-pipeline"
  description     = "API key for CI/CD pipeline"
  scopes          = ["modules:read", "modules:write", "providers:read", "providers:write"]
  expires_at      = "2027-01-01T00:00:00Z"
}

output "ci_api_key" {
  value     = registry_api_key.ci.key
  sensitive = true
}
