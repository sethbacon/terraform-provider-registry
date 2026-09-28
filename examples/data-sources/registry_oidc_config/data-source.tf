data "registry_oidc_config" "current" {}

output "oidc_issuer" {
  value = data.registry_oidc_config.current.issuer_url
}
