data "registry_stats" "dashboard" {}

output "total_modules" {
  value = data.registry_stats.dashboard.total_modules
}

output "total_providers" {
  value = data.registry_stats.dashboard.total_providers
}
