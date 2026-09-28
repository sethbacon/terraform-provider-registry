data "registry_terraform_mirror_history" "opentofu" {
  mirror_id = registry_terraform_mirror.opentofu.id
}

output "last_sync_status" {
  value = try(data.registry_terraform_mirror_history.opentofu.history[0].status, null)
}
