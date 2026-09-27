# One-shot: create starts the migration and waits for a terminal status.
# Changing either config ID replaces the resource, which starts a new migration.
resource "registry_storage_migration" "s3_to_azure" {
  source_config_id = registry_storage_config.s3.id
  target_config_id = registry_storage_config.azure.id
  timeout_minutes  = 120
}
