resource "registry_terraform_mirror" "opentofu" {
  name                = "OpenTofu Releases"
  tool                = "opentofu"
  upstream_url        = "https://get.opentofu.org/tofu"
  platform_filter     = ["linux/amd64", "linux/arm64"]
  gpg_verify          = true
  stable_only         = true
  sync_interval_hours = 24
}
