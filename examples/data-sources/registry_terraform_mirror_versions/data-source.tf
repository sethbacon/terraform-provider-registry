data "registry_terraform_mirror_versions" "opentofu" {
  mirror_id = registry_terraform_mirror.opentofu.id
}

output "stable_versions" {
  value = [for v in data.registry_terraform_mirror_versions.opentofu.versions : v.version if v.stable]
}
