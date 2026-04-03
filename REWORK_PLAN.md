# Vaultbot Lite v2 — Rework Plan

## Overview

Simplify the project from a long-running Discord bot to a set of stateless GitHub Actions cron jobs.
The Neon PostgreSQL database is preserved as the source of truth for track history and rankings.
Tracks are added directly to the Spotify playlist by the user; a poll job detects them and records them in the DB.

---

## What Gets Removed

### Application code
- `cmd/vaultbot/` — main bot entrypoint
- `cmd/migration_runner/` — standalone migration binary (migrations move to a GH Actions workflow)
- `internal/discord/` — entire Discord integration (bot, commands, helpers)
- `internal/blacklist/` — blacklist feature
- `internal/preferences/` — preferences feature and dynamic scheduling
- `internal/tracks/discord.go` — Discord-specific track add handler
- In-memory track cache (only needed for the always-on bot)
- Health check HTTP server

### Infrastructure
- `infra/` — all Terraform (was managing DigitalOcean App Platform only; Neon was created manually)
- `.github/workflows/terraform.yml`
- `Vaultbot.Dockerfile`
- `MigrationRunner.Dockerfile`
- `docker-compose.yml`

---

## What Stays (Largely Intact)

- `internal/spotify/` — Spotify client, trimmed of unused methods and the interactive OAuth flow
- `internal/persistence/postgres/` — full DB layer and all existing migrations
- `internal/domain/` — core models, stripped of blacklist/preference types
- Genre playlist population logic (`internal/cron/populate_genre_playlist.go`)
- High scores playlist population logic (`internal/cron/populate_high_scores_playlist.go`)
- Purge logic (`internal/cron/purge_tracks.go`)
- `internal/utils/token.go` — Spotify token parsing (oauth2 auto-refresh, no changes needed)
- `.github/workflows/go.yml` — CI build and test

---

## Database Changes (new migrations)

### Migration: drop unused tables
- Drop `blacklist` table
- Drop `preferences` table

### Migration: remove user attribution
- Drop `users` table
- Drop `song_archive.user_id` column (and its FK constraint)

> Existing `song_archive` rows are preserved — only the user attribution column is removed.
> The `added_by` Spotify user ID is available from the playlist items API but there is no current
> need to store it; this can be revisited later.

---

## New Architecture: GitHub Actions Jobs

Each job is a small standalone `cmd/` binary that runs, does its work, and exits.
All jobs share the same environment secrets.

### Job 1 — `poll-playlist`
**Schedule:** 4x daily (every 6 hours)
**Binary:** `cmd/poll/`

Reads all current items from the main Spotify playlist. For each track, checks whether a
`song_archive` entry exists with `created_at >= track's Spotify added_at`. If not, the track
is a new addition event:
1. Upsert track metadata into `songs`
2. Upsert artist/genre relationships
3. Insert a new `song_archive` row

Handles re-added tracks correctly (a track purged and re-added gets a new `song_archive` entry,
which is correct for ranking purposes).

### Job 2 — `purge-tracks`
**Schedule:** Every 12 hours (or daily)
**Binary:** `cmd/purge/`

Reads all current items from the main Spotify playlist. Removes any track where Spotify's
`added_at` is older than 2 weeks. Hardcoded — not configurable.

### Job 3 — `genre-playlist`
**Schedule:** Daily at 00:00 UTC
**Binary:** `cmd/genre/`

Queries the DB for a random genre's tracks, computes a diff against the current genre playlist,
and updates the Spotify playlist + description. Logic ported directly from
`internal/cron/populate_genre_playlist.go`.

### Job 4 — `high-scores-playlist`
**Schedule:** Daily at 00:00 UTC
**Binary:** `cmd/highscores/`

Queries the DB for the top 50 tracks (by `song_archive` frequency), computes a diff against the
current high scores playlist, and updates it. Logic ported directly from
`internal/cron/populate_high_scores_playlist.go`.

---

## New Architecture: Migration Workflow

**Trigger:** Push to `main` (scoped to changes under `internal/persistence/postgres/migrations/`)
**Also:** Manual `workflow_dispatch` for one-off runs

Runs the migration runner logic (currently in `cmd/migration_runner/`) as a GH Actions step.
The existing idempotency check (skip already-run migrations) is preserved.

---

## Spotify Authentication

No changes required. The existing `SPOTIFY_TOKEN` env var (pipe-delimited
`accessToken|refreshToken|tokenType|expiryUnix`) is stored as a GitHub Actions secret.
The `golang.org/x/oauth2` transport automatically refreshes the access token on each job run
using the embedded refresh token. Spotify refresh tokens do not expire unless revoked.

### Required GitHub Actions Secrets
| Secret | Purpose |
|---|---|
| `SPOTIFY_TOKEN` | Full token string (existing format) |
| `SPOTIFY_CLIENT_ID` | Spotify app client ID |
| `SPOTIFY_CLIENT_SECRET` | Spotify app client secret |
| `SPOTIFY_PLAYLIST_ID` | Main dynamic playlist |
| `GENRE_SPOTIFY_PLAYLIST_ID` | Genre rotation playlist |
| `HIGH_SCORES_SPOTIFY_PLAYLIST_ID` | Top 50 playlist |
| `DB_*` | Neon connection config (existing vars) |

---

## Target Project Structure

```
cmd/
  poll/           # Detects new playlist additions, writes to DB
  purge/          # Removes expired tracks from playlist
  genre/          # Repopulates genre playlist
  highscores/     # Repopulates top 50 playlist
internal/
  spotify/        # Trimmed Spotify client (no interactive OAuth)
  persistence/    # Full DB layer (unchanged)
  domain/         # Core models (blacklist/preference types removed)
  cron/           # Playlist population logic (reused by genre + highscores cmds)
  utils/          # token.go and other retained utilities
.github/
  workflows/
    ci.yml              # go build + test (existing go.yml, renamed)
    poll.yml
    purge.yml
    curated-playlists.yml
    migrations.yml
```

---

## Implementation Order

1. New DB migrations (drop blacklist, preferences, users, song_archive.user_id)
2. Strip Discord, blacklist, preferences from application code
3. Remove interactive OAuth flow from `internal/spotify/spotify.go`
4. Refactor existing cron logic into importable packages (decouple from gocron)
5. Write `cmd/poll/`, `cmd/purge/`, `cmd/genre/`, `cmd/highscores/`
6. Write GitHub Actions workflow files
7. Delete Terraform, Dockerfiles, docker-compose
8. Update CI workflow (`go.yml`)
