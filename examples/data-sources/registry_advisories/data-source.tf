data "registry_advisories" "active_critical" {
  active_only = true
  severities  = ["high", "critical"]
}

output "active_cves" {
  value = [for a in data.registry_advisories.active_critical.advisories : a.external_id]
}
