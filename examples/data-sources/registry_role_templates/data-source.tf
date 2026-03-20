data "registry_role_templates" "all" {}

output "role_template_names" {
  value = [for r in data.registry_role_templates.all.role_templates : r.name]
}

output "admin_role_id" {
  value = one([for r in data.registry_role_templates.all.role_templates : r.id if r.name == "admin"])
}
