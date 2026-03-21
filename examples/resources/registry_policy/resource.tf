resource "registry_policy" "allow_hashicorp" {
  name              = "allow-hashicorp"
  description       = "Allow mirroring hashicorp providers"
  policy_type       = "allow"
  namespace_pattern = "hashicorp/*"
  is_active         = true
  requires_approval = false
}
