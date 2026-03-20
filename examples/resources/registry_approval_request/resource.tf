resource "registry_approval_request" "hashicorp_mirror" {
  mirror_id      = registry_mirror.hashicorp.id
  justification  = "Need to sync HashiCorp providers for production deployment"
}
