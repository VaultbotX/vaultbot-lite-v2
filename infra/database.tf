resource "digitalocean_database_cluster" "vaultbot_postgres_cluster" {
  name       = "vaultbot-postgres-cluster-${var.environment}"
  engine     = "pg"
  version    = "17"
  size       = "db-s-1vcpu-1gb"
  region     = var.do_region
  node_count = 1
  private_network_uuid = digitalocean_vpc.vaultbot_vpc.id

  maintenance_window {
    day  = "monday"
    hour = "03:00"
  }
}