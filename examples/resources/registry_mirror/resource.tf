resource "registry_mirror" "hashicorp" {
  name                  = "HashiCorp Providers"
  description           = "Mirror of HashiCorp official providers"
  upstream_registry_url = "https://registry.terraform.io"
  organization_id       = registry_organization.example.id
  namespace_filter      = ["hashicorp"]
  provider_filter       = ["aws", "azurerm", "google"]
  platform_filter       = ["linux/amd64", "linux/arm64", "darwin/arm64"]
  enabled               = true
  sync_interval_hours   = 24
}
