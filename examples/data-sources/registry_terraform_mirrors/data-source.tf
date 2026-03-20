data "registry_terraform_mirrors" "all" {}

output "terraform_mirror_names" {
  value = [for m in data.registry_terraform_mirrors.all.terraform_mirrors : m.name]
}
