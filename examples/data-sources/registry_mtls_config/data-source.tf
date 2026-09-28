data "registry_mtls_config" "current" {}

output "mtls_enabled" {
  value = data.registry_mtls_config.current.enabled
}
