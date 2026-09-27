data "registry_policy_evaluation" "mirror_upstream" {
  query = "data.authz.allow"
  input_json = jsonencode({
    action   = "mirror"
    upstream = "registry.terraform.io"
  })
}

output "mirror_allowed" {
  value = data.registry_policy_evaluation.mirror_upstream.allowed
}
