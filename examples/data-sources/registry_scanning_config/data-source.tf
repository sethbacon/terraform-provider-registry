data "registry_scanning_config" "current" {}

output "scanning_enabled" {
  value = data.registry_scanning_config.current.enabled
}
