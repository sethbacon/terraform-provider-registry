resource "registry_organization_member" "example" {
  organization_id  = registry_organization.example.id
  user_id          = registry_user.example.id
  role_template_id = data.registry_role_templates.all.role_templates[0].id
}
