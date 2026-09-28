data "registry_module_scan" "vpc" {
  namespace = "my-org"
  name      = "vpc"
  system    = "aws"
  version   = "1.2.0"
}

output "vpc_scan_passed" {
  value = data.registry_module_scan.vpc.passed
}
