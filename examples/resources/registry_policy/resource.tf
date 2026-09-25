# Advisory only: the registry records this policy and reports it from its
# dry-run evaluation, but mirror sync and pull-through do not enforce it.
# Restrict what a mirror fetches with the registry_mirror filters instead.
resource "registry_policy" "allow_hashicorp" {
  name              = "allow-hashicorp"
  description       = "Allow mirroring hashicorp providers"
  policy_type       = "allow"
  namespace_pattern = "hashicorp"
  is_active         = true
  requires_approval = false
}
