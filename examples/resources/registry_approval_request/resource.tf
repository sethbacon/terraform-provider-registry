# Advisory only: the registry records the request and its review status, but
# mirror sync and pull-through do not check it.
resource "registry_approval_request" "hashicorp_mirror" {
  mirror_id          = registry_mirror.hashicorp.id
  provider_namespace = "hashicorp"
  justification      = "Need to sync HashiCorp providers for production deployment"
}
