resource "digitalocean_database_cluster" "vaultbot_postgres_cluster" {
  name       = "vaultbot-postgres-cluster-${var.environment}"
  engine     = "pg"
  version    = "17"
  size       = "db-s-1vcpu-1gb"
  region     = var.do_region
  node_count = 1

  maintenance_window {
    day  = "monday"
    hour = "03:00"
  }
}

resource "digitalocean_database_firewall" "vaultbot-postgres-firewall" {
  cluster_id = digitalocean_database_cluster.vaultbot_postgres_cluster.id
  # no inbound rules configured here. will try to manage this via the UI
}