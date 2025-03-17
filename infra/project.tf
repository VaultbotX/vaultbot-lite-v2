resource "digitalocean_project" "vaultbot" {
  name        = "Vaultbot-${var.environment}"
  description = "Vaultbot Project for ${var.environment} environment"
  resources = [
    digitalocean_database_cluster.vaultbot_postgres_cluster.urn,
    digitalocean_app.vaultbot_app.urn,
    digitalocean_app.vaultbot_migration_runner.urn
  ]
}