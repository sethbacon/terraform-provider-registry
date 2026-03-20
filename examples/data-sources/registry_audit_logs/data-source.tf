data "registry_audit_logs" "recent" {
  limit = 50
}

data "registry_audit_logs" "module_events" {
  resource_type = "module"
  action        = "create"
  limit         = 100
}

output "recent_log_count" {
  value = data.registry_audit_logs.recent.total
}
