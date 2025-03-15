resource "digitalocean_app" "vaultbot_app" {
  depends_on = [digitalocean_database_cluster.vaultbot_postgres_cluster]
  spec {
    name   = "vaultbot-app-${var.environment}"
    region = var.do_region

    database {
      name         = "vaultbot-postgres-${var.environment}"
      engine       = "pg"
      cluster_name = digitalocean_database_cluster.vaultbot_postgres_cluster.name
      db_name      = digitalocean_database_db.vaultbot_db.name
    }

    service {
      name               = "websocket-backend-${var.environment}"
      instance_count     = 1
      instance_size_slug = "basic-xxs"
      source_dir         = "."
      dockerfile_path    = "Vaultbot.Dockerfile"

      autoscaling {
        min_instance_count = 1
        max_instance_count = 2

        metrics {
          cpu {
            percent = 70
          }
        }
      }

      env {
        key   = "POSTGRES_HOST"
        value = digitalocean_database_connection_pool.vaultbot_pool.host
        scope = "RUN_TIME"
        type = "SECRET"
      }

      env {
        key   = "POSTGRES_USER"
        value = digitalocean_database_connection_pool.vaultbot_pool.user
        scope = "RUN_TIME"
        type = "SECRET"
      }

      env {
        key   = "POSTGRES_PASSWORD"
        value = digitalocean_database_user.vaultbot_user.password
        scope = "RUN_TIME"
        type = "SECRET"
      }

      env {
        key = "DISCORD_TOKEN"
        value = var.discord_token
        scope = "RUN_TIME"
        type = "SECRET"
      }

      env {
        key = "SPOTIFY_PLAYLIST_ID"
        value = var.spotify_playlist_id
        scope = "RUN_TIME"
        type = "SECRET"
      }

      env {
          key = "SPOTIFY_CLIENT_ID"
          value = var.spotify_client_id
          scope = "RUN_TIME"
          type = "SECRET"
      }

      env {
          key = "SPOTIFY_CLIENT_SECRET"
          value = var.spotify_client_secret
          scope = "RUN_TIME"
          type = "SECRET"
      }

      env {
        key = "SPOTIFY_TOKEN"
        value = var.spotify_token
        scope = "RUN_TIME"
        type = "SECRET"
      }

      github {
        repo           = var.github_repo
        branch         = var.github_repo_branch
        deploy_on_push = true
      }
    }

    job {
      name       = "migration-runner-${var.environment}"
      source_dir = "."
      dockerfile_path = "MigrationRunner.Dockerfile"

      env {
        key   = "POSTGRES_HOST"
        value = digitalocean_database_connection_pool.vaultbot_pool.host
        scope = "RUN_TIME"
        type = "SECRET"
      }

      env {
        key   = "POSTGRES_USER"
        value = digitalocean_database_connection_pool.vaultbot_pool.user
        scope = "RUN_TIME"
        type = "SECRET"
      }

      env {
        key   = "POSTGRES_PASSWORD"
        value = digitalocean_database_user.vaultbot_user.password
        scope = "RUN_TIME"
        type = "SECRET"
      }

      github {
        repo           = var.github_repo
        branch         = var.github_repo_branch
        deploy_on_push = true
      }
    }
  }
}