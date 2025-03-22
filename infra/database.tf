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

resource "digitalocean_database_postgresql_config" "vaultbot_postgres_config" {
  cluster_id = digitalocean_database_cluster.vaultbot_postgres_cluster.id
  timezone   = "UTC"
}

resource "digitalocean_database_db" "vaultbot_db" {
  cluster_id = digitalocean_database_cluster.vaultbot_postgres_cluster.id
  // Application code currently requires a database named "vaultbot"
  name       = "vaultbot"
}

resource "digitalocean_database_firewall" "vaultbot_postgres_firewall" {
  cluster_id = digitalocean_database_cluster.vaultbot_postgres_cluster.id

  rule {
    type  = "app"
    value = digitalocean_app.vaultbot_migration_runner.id
  }

  rule {
    type  = "app"
    value = digitalocean_app.vaultbot_app.id
  }
}
