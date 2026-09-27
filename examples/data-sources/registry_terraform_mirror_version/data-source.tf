data "registry_terraform_mirror_version" "tofu" {
  mirror_id = registry_terraform_mirror.opentofu.id
  version   = "1.10.0"
}

output "tofu_platforms" {
  value = [for p in data.registry_terraform_mirror_version.tofu.platforms : "${p.os}/${p.arch}"]
}
