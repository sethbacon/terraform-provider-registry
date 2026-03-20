resource "registry_module" "vpc" {
  organization_id = registry_organization.example.id
  namespace       = "my-org"
  name            = "vpc"
  system          = "aws"
  description     = "AWS VPC module"
  source          = "https://github.com/my-org/terraform-aws-vpc"
}
