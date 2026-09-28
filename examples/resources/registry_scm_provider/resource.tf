# Set organization_id: backend 4.18 and later refuse a platform-admin create
# of an SCM provider that names no organization.
resource "registry_scm_provider" "github" {
  organization_id = registry_organization.example.id
  name            = "GitHub"
  type            = "github"
  client_id       = var.github_oauth_client_id
  client_secret   = var.github_oauth_client_secret
}

resource "registry_scm_provider" "self_hosted_gitlab" {
  organization_id = registry_organization.example.id
  name            = "Internal GitLab"
  type            = "gitlab"
  base_url        = "https://gitlab.mycompany.com"
  client_id       = var.gitlab_oauth_client_id
  client_secret   = var.gitlab_oauth_client_secret
}
