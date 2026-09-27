data "registry_audit_log" "entry" {
  id = "00000000-0000-0000-0000-000000000000"
}

output "audit_actor" {
  value = data.registry_audit_log.entry.user_email
}
