terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "digitalocean" {
  token = var.do_management_token
}

resource "digitalocean_project" "vaultbot" {
  name        = "Vaultbot-${var.environment}"
  description = "Vaultbot Project"
  resources = [
    digitalocean_database_cluster.vaultbot_postgres_cluster.id,
    digitalocean_app.vaultbot_app.id
  ]
}

resource "digitalocean_database_cluster" "vaultbot_postgres_cluster" {
  name       = "vaultbot-postgres-cluster-${var.environment}"
  engine     = "pg"
  version    = "17"
  size       = "db-s-1vcpu-1gb"
  region     = var.do_region
  node_count = 1
}

resource "digitalocean_app" "vaultbot_app" {
  depends_on = [digitalocean_database_cluster.vaultbot_postgres_cluster]
  spec {
    name   = "vaultbot-app-${var.environment}"
    region = var.do_region

    database {
      name         = "vaultbot-postgres-${var.environment}"
      engine       = "pg"
      cluster_name = digitalocean_database_cluster.vaultbot_postgres_cluster.name
      db_name      = "vaultbot"
      user         = "vaultbot"
    }

    service {
      name               = "websocket-backend-${var.environment}"
      instance_count     = 1
      instance_size_slug = "basic-xxs"
      source_dir         = "."
      dockerfile_path    = "Vaultbot.Dockerfile"

      env {
        key   = "DATABASE_URL"
        value = digitalocean_database_cluster.vaultbot_postgres_cluster.uri
        scope = "RUN_TIME"
        type = "SECRET"
      }

      github {
        repo           = var.github_repo_url
        branch         = var.github_repo_branch
        deploy_on_push = true
      }
    }

    job {
      name       = "migration-runner-${var.environment}"
      source_dir = "."
      dockerfile_path = "MigrationRunner.Dockerfile"

      env {
        key   = "DATABASE_URL"
        value = digitalocean_database_cluster.vaultbot_postgres_cluster.uri
        scope = "RUN_TIME"
        type = "SECRET"
      }

      github {
        repo           = var.github_repo_url
        branch         = var.github_repo_branch
        deploy_on_push = true
      }
    }
  }
}