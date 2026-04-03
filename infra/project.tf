resource "digitalocean_project" "vaultbot" {
  name        = "Vaultbot-${var.environment}"
  description = "Vaultbot Project for ${var.environment} environment"
  resources = [
    digitalocean_app.vaultbot_app.urn,
    digitalocean_app.vaultbot_migration_runner.urn
  ]
}