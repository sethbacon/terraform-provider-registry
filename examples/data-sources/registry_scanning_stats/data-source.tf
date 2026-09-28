data "registry_scanning_stats" "current" {}

output "critical_findings" {
  value = data.registry_scanning_stats.current.critical_count
}
