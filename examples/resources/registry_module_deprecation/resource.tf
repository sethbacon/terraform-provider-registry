resource "registry_module_deprecation" "legacy_vpc" {
  namespace           = "my-org"
  name                = "legacy-vpc"
  system              = "aws"
  message             = "Use my-org/vpc/aws instead."
  successor_module_id = registry_module.vpc.id
}
