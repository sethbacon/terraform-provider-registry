resource "registry_scm_provider" "github" {
  name = "GitHub"
  type = "github"
}

resource "registry_scm_provider" "self_hosted_gitlab" {
  name     = "Internal GitLab"
  type     = "gitlab"
  base_url = "https://gitlab.mycompany.com"
}
