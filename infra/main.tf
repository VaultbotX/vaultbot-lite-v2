terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

# Set the variable value in *.tfvars file
# or using -var="do_token=..." CLI option
variable "do_token" {}

variable "do_region" {
  type        = string
  description = "Region for the resources"
}

variable "github_repo_url" {
  type        = string
  description = "GitHub repository URL for the app"
}

variable "github_repo_branch" {
  type        = string
  description = "Branch to deploy from"
}

# Configure the DigitalOcean Provider
provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_project" "vaultbot" {
  name        = "Vaultbot"
  description = "Vaultbot Project"
  resources = [
    digitalocean_database_cluster.vaultbot_postgres_cluster.id,
    digitalocean_app.vaultbot_app.id
  ]
}

resource "digitalocean_database_cluster" "vaultbot_postgres_cluster" {
  name       = "vaultbot-postgres-cluster"
  engine     = "pg"
  version    = "17"
  size       = "db-s-1vcpu-1gb"
  region     = var.do_region
  node_count = 1
}

# DigitalOcean App Platform Deployment
resource "digitalocean_app" "vaultbot_app" {
  depends_on = [digitalocean_database_cluster.vaultbot_postgres_cluster]
  spec {
    name   = "vaultbot-app"
    region = var.do_region

    # Postgres database as a component in the app
    database {
      name         = "vaultbot-postgres"
      engine       = "pg"
      cluster_name = digitalocean_database_cluster.vaultbot_postgres_cluster.name
      db_name      = "vaultbot"
      user         = "vaultbot"
    }

    # Go WebSocket Backend Service
    service {
      name               = "websocket-backend"
      instance_count     = 1
      instance_size_slug = "basic-xxs"
      source_dir         = "."
      dockerfile_path    = "Vaultbot.Dockerfile"

      env {
        key   = "DATABASE_URL"
        value = digitalocean_database_cluster.vaultbot_postgres_cluster.uri
        scope = "RUN_TIME"
      }

      github {
        repo           = var.github_repo_url
        branch         = var.github_repo_branch
        deploy_on_push = true
      }
    }

    # Job for Database Migrations
    job {
      name       = "migration-runner"
      source_dir = "."
      dockerfile_path = "MigrationRunner.Dockerfile"

      # Runs once per deployment and exits
      run_command = "your-migration-command"

      env {
        key   = "DATABASE_URL"
        value = digitalocean_database_cluster.vaultbot_postgres_cluster.uri
        scope = "RUN_TIME"
      }

      github {
        repo           = var.github_repo_url
        branch         = var.github_repo_branch
        deploy_on_push = true
      }
    }
  }
}