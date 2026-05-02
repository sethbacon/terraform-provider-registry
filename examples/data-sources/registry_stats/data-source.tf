data "registry_stats" "dashboard" {}

output "total_modules" {
  value = data.registry_stats.dashboard.modules.total
}

output "total_providers" {
  value = data.registry_stats.dashboard.providers.total
}

output "total_users" {
  value = data.registry_stats.dashboard.users
}

output "manual_provider_versions" {
  value = data.registry_stats.dashboard.providers.manual_versions
}

output "binary_mirrors_healthy" {
  value = data.registry_stats.dashboard.binary_mirrors.healthy
}
