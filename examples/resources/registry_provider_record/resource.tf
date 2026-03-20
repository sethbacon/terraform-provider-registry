resource "registry_provider_record" "example" {
  organization_id = registry_organization.example.id
  namespace       = "my-org"
  type            = "example"
  description     = "Example Terraform provider"
  source          = "https://github.com/my-org/terraform-provider-example"
}
