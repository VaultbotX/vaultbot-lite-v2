resource "digitalocean_app" "vaultbot_app" {
  depends_on = [digitalocean_app.vaultbot_migration_runner]
  spec {
    name   = "vaultbot-app-${var.environment}"
    region = var.do_region

    alert {
      rule = "DEPLOYMENT_FAILED"
    }

    alert {
      rule = "DEPLOYMENT_CANCELED"
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
        http_path             = "/api/healthz"
        port                  = 8080
        period_seconds        = 10
        timeout_seconds       = 3
        failure_threshold     = 3
        success_threshold     = 1
        initial_delay_seconds = 5
      }

      env {
        key   = "POSTGRES_HOST"
        value = var.neon_host
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "POSTGRES_PORT"
        value = var.neon_port
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "POSTGRES_USER"
        value = var.neon_user
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "POSTGRES_PASSWORD"
        value = var.neon_password
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "POSTGRES_DB"
        value = var.neon_db
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
        key   = "DISCORD_ADMINISTRATOR_USER_ID"
        value = var.discord_administrator_user_id
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
        key   = "GENRE_SPOTIFY_PLAYLIST_ID"
        value = var.genre_spotify_playlist_id
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "HIGH_SCORES_SPOTIFY_PLAYLIST_ID"
        value = var.high_scores_spotify_playlist_id
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

      env {
        key   = "APP_VERSION"
        value = var.app_version
        scope = "RUN_TIME"
        type  = "GENERAL"
      }

      github {
        repo           = var.github_repo
        branch         = var.github_repo_branch
        deploy_on_push = true
      }
    }
  }
}

resource "digitalocean_app" "vaultbot_migration_runner" {
  spec {
    name   = "vaultbot-migration-runner-${var.environment}"
    region = var.do_region

    alert {
      rule = "DEPLOYMENT_FAILED"
    }

    alert {
      rule = "DEPLOYMENT_CANCELED"
    }

    job {
      name               = "migration-runner-${var.environment}"
      instance_count     = 1
      instance_size_slug = "basic-xxs"
      source_dir         = "."
      dockerfile_path    = "MigrationRunner.Dockerfile"

      env {
        key   = "POSTGRES_HOST"
        value = var.neon_host
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "POSTGRES_PORT"
        value = var.neon_port
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "POSTGRES_USER"
        value = var.neon_user
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "POSTGRES_PASSWORD"
        value = var.neon_password
        scope = "RUN_TIME"
        type  = "SECRET"
      }

      env {
        key   = "POSTGRES_DB"
        value = var.neon_db
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
