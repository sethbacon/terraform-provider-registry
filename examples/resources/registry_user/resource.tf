resource "registry_user" "example" {
  email = "jane.doe@example.com"
  name  = "Jane Doe"
}

resource "registry_user" "oidc" {
  email    = "john.smith@example.com"
  name     = "John Smith"
  oidc_sub = "auth0|abc123xyz"
}
