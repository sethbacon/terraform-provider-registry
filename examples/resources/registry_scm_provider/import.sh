# An SCM provider is imported by its UUID. The API never returns client_secret or
# webhook_secret, so set them in configuration after import.
terraform import registry_scm_provider.github 00000000-0000-0000-0000-000000000000
