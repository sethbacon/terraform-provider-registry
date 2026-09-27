resource "registry_module_version_deprecation" "vpc_1_0_0" {
  namespace          = "my-org"
  name               = "vpc"
  system             = "aws"
  version            = "1.0.0"
  message            = "1.0.0 creates public subnets by default; upgrade to 1.2.0 or later."
  replacement_source = "registry.example.com/my-org/vpc/aws"
}
