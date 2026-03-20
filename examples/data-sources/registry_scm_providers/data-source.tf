data "registry_scm_providers" "all" {}

output "scm_names" {
  value = [for s in data.registry_scm_providers.all.scm_providers : s.name]
}
