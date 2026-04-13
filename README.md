# vaultbot-lite-v2

Catalogs music listening by managing a Spotify playlist. Tracks added directly to the playlist are detected and recorded in a Neon PostgreSQL database. Tracks older than two weeks are automatically purged. Two curated playlists — a rotating genre selection and a top-50 all-time chart — are refreshed daily.

All jobs run as stateless GitHub Actions cron jobs. There is no long-running service.

## How it works

| Workflow | Schedule | Description |
|---|---|---|
| **Poll Playlist** | Every 6 hours | Detects tracks manually added to the main playlist and records them in the DB |
| **Purge Expired Tracks** | Twice daily | Removes tracks older than 2 weeks from the main playlist |
| **Curated Playlists** | Daily at midnight UTC | Refreshes the genre rotation playlist and the top-50 high scores playlist |
| **Run Migrations** | On push to `main` (migration files only) | Runs any new database migrations against Neon |

All workflows can also be triggered manually from the GitHub Actions UI via `workflow_dispatch`.

## Requirements

- Go 1.26
- A [Neon](https://neon.tech) PostgreSQL database
- A Spotify Developer application

## Configuration

All required environment variables are documented in [`.env.example`](.env.example). Copy it to `.env` for local development — `godotenv` loads it automatically.

In GitHub Actions, variables are stored as repository secrets. See [`.devcontainer/SETUP.md`](.devcontainer/SETUP.md) for the complete list of secret names as they must be configured.

### Spotify token

`SPOTIFY_TOKEN` is a serialized OAuth2 token in the format `accessToken|refreshToken|tokenType|expiryUnix`. It must be obtained via Spotify's Authorization Code flow using the scopes listed below, then stored as a secret. The `golang.org/x/oauth2` library automatically refreshes it on each run using the embedded refresh token, so it never needs to be manually updated after the initial setup.

Required scopes: `playlist-modify-public`, `playlist-modify-private`, `playlist-read-private`, `playlist-read-collaborative`

## GitHub Codespaces

This repo includes a devcontainer for use with GitHub Codespaces. When a codespace starts, it automatically creates a Neon database branch scoped to the current git branch and writes the connection details into `.env` — no local Postgres container needed.

See [`.devcontainer/SETUP.md`](.devcontainer/SETUP.md) for first-time setup instructions.

When finished with a branch, delete its Neon branch by running:

```sh
bash .devcontainer/scripts/neon-branch-teardown.sh
```

## Database schema

![db schema](assets/schema.png "schema")
