resource "registry_provider_version_deprecation" "example_0_9" {
  namespace = "my-org"
  type      = "example"
  version   = "0.9.0"
  message   = "0.9.0 has a known state-corruption bug; upgrade to 0.9.1 or later."
}
