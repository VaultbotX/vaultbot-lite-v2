resource "digitalocean_app" "vaultbot_app" {
  depends_on = [digitalocean_database_cluster.vaultbot_postgres_cluster]
  spec {
    name   = "vaultbot-app-${var.environment}"
    region = var.do_region

    alert {
      rule = "DEPLOYMENT_FAILED"
    }

    alert {
      rule = "DEPLOYMENT_CANCELED"
    }

    alert {
      rule = "DEPLOYMENT_LIVE"
    }

    service {
      name               = "websocket-backend-${var.environment}"
      instance_count     = 1
      instance_size_slug = "basic-xxs"
      source_dir         = "."
      dockerfile_path    = "Vaultbot.Dockerfile"

      alert {
        operator = "GREATER_THAN"
        rule     = "RESTART_COUNT"
        value    = 2
        window   = "TEN_MINUTES"
      }

      health_check {
        protocol = "HTTP"
        path     = "/healthz"
        port     = 8080
        check_interval_seconds = 60
        response_timeout_seconds = 5
        healthy_threshold = 3
        unhealthy_threshold = 3
      }

      env {
        key   = "POSTGRES_HOST"
        value = digitalocean_database_connection_pool.vaultbot_pool.host
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "POSTGRES_USER"
        value = digitalocean_database_connection_pool.vaultbot_pool.user
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "POSTGRES_PASSWORD"
        value = digitalocean_database_user.vaultbot_user.password
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "DISCORD_TOKEN"
        value = var.discord_token
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "SPOTIFY_PLAYLIST_ID"
        value = var.spotify_playlist_id
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "SPOTIFY_CLIENT_ID"
        value = var.spotify_client_id
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "SPOTIFY_CLIENT_SECRET"
        value = var.spotify_client_secret
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "SPOTIFY_TOKEN"
        value = var.spotify_token
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      github {
        repo           = var.github_repo
        branch         = var.github_repo_branch
        deploy_on_push = true
      }
    }

    job {
      name            = "migration-runner-${var.environment}"
      source_dir      = "."
      dockerfile_path = "MigrationRunner.Dockerfile"

      env {
        key   = "POSTGRES_HOST"
        value = digitalocean_database_connection_pool.vaultbot_pool.host
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "POSTGRES_USER"
        value = digitalocean_database_connection_pool.vaultbot_pool.user
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "POSTGRES_PASSWORD"
        value = digitalocean_database_user.vaultbot_user.password
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      github {
        repo           = var.github_repo
        branch         = var.github_repo_branch
        deploy_on_push = true
      }
    }
  }
}