resource "digitalocean_project" "vaultbot" {
  name        = "Vaultbot-${var.environment}"
  description = "Vaultbot Project for ${var.environment} environment"
  resources = [
    digitalocean_database_cluster.vaultbot_postgres_cluster.id,
    digitalocean_app.vaultbot_app.id
  ]
}