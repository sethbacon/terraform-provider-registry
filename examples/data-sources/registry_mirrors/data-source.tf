data "registry_mirrors" "all" {}

output "mirror_names" {
  value = [for m in data.registry_mirrors.all.mirrors : m.name]
}
