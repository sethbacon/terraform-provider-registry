data "registry_identity_group_mappings" "current" {}

output "mapped_groups" {
  value = [for m in data.registry_identity_group_mappings.current.mappings : m.group]
}
