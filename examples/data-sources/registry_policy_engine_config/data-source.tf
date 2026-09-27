data "registry_policy_engine_config" "current" {}

output "policy_bundle_status" {
  value = data.registry_policy_engine_config.current.status
}
