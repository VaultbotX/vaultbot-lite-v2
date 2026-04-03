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

The following environment variables are required. In GitHub Actions they are stored as repository secrets. For local development, set them in a `.env` file in the root directory and also set `ENVIRONMENT=local` (any non-empty string) to disable SSL on the database connection.

### Spotify

| Variable | Description |
|---|---|
| `SPOTIFY_CLIENT_ID` | Spotify application client ID |
| `SPOTIFY_CLIENT_SECRET` | Spotify application client secret |
| `SPOTIFY_TOKEN` | Serialized OAuth token (see below) |
| `SPOTIFY_PLAYLIST_ID` | ID of the main dynamic playlist |
| `GENRE_SPOTIFY_PLAYLIST_ID` | ID of the genre rotation playlist |
| `HIGH_SCORES_SPOTIFY_PLAYLIST_ID` | ID of the top-50 playlist |

### Database

| Variable | Description |
|---|---|
| `POSTGRES_HOST` | Neon hostname |
| `POSTGRES_PORT` | Neon port |
| `POSTGRES_USER` | Neon user |
| `POSTGRES_PASSWORD` | Neon password |
| `POSTGRES_DB` | Database name (defaults to `vaultbot` if unset) |

### Spotify token setup

The Spotify token is required because some playlist write scopes are only available via the Authorization Code flow, not Client Credentials.

To generate the token for the first time:

1. Register `http://localhost:8080/callback` as a redirect URI in the Spotify Developer Dashboard
2. Run the app locally without `SPOTIFY_TOKEN` set — a browser window will open for Spotify login
3. After authenticating, a `token.txt` file is created in the project root
4. Copy its contents and store them as the `SPOTIFY_TOKEN` secret

The token string contains a refresh token. The `golang.org/x/oauth2` library automatically exchanges it for a fresh access token on each run, so the secret value never needs to be updated.

> **Note:** The audio features endpoint is deprecated and not used.
> See: https://developer.spotify.com/blog/2024-11-27-changes-to-the-web-api

## Database schema

![db schema](assets/schema.png "schema")
