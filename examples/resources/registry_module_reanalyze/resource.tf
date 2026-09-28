# Re-runs analysis (terraform-docs, scanning, SCM verification) for a module
# version. Change a value in triggers to run it again.
resource "registry_module_reanalyze" "vpc_1_2_0" {
  namespace = "my-org"
  name      = "vpc"
  system    = "aws"
  version   = "1.2.0"

  triggers = {
    scanner_version = "0.58.0"
  }
}
