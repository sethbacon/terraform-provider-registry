resource "registry_module_scm_link" "vpc" {
  module_id      = registry_module.vpc.id
  scm_provider_id = registry_scm_provider.github.id
  owner          = "my-org"
  repo           = "terraform-aws-vpc"
  branch         = "main"
  tag_pattern    = "v*"
}
