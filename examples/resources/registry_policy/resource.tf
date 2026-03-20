resource "registry_policy" "require_approval" {
  name        = "require-approval"
  description = "Requires approval before mirror sync"
  rules       = jsonencode({
    require_approval = true
    approvers        = ["admin"]
  })
}
