data "registry_users" "all" {}

data "registry_users" "search" {
  search = "jane"
}

output "user_emails" {
  value = [for u in data.registry_users.all.users : u.email]
}
