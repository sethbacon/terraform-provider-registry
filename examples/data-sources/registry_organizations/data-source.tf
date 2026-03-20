data "registry_organizations" "all" {}

output "org_names" {
  value = [for o in data.registry_organizations.all.organizations : o.name]
}
