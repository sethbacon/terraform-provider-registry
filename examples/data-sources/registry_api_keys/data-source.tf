data "registry_api_keys" "all" {}

data "registry_api_keys" "user_keys" {
  user_id = registry_user.example.id
}
