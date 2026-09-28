# Each mapping is one entry in the backend's full-replace list. Manage every
# mapping for an organization through this resource so that out-of-band edits
# do not race with Terraform's read-modify-write.
resource "registry_oidc_group_mapping" "platform_admins" {
  group        = "platform-admins"
  organization = registry_organization.example.name
  role         = "admin"
}
