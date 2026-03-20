resource "registry_role_template" "module_publisher" {
  name         = "module-publisher"
  display_name = "Module Publisher"
  description  = "Can publish and manage modules"
  scopes       = ["modules:read", "modules:write", "organizations:read"]
}
