data "registry_providers" "all" {}

data "registry_providers" "by_namespace" {
  namespace = "my-org"
}
