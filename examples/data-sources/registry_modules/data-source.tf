data "registry_modules" "all" {}

data "registry_modules" "by_namespace" {
  namespace = "my-org"
}

data "registry_modules" "search" {
  search = "vpc"
}
