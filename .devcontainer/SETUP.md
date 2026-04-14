# First-time setup guide

This document covers everything that must be configured manually in GitHub, Neon, and Spotify before the workflows and Codespaces will work.

---

## 1. Spotify

### 1a. Create a Developer application

1. Go to [developer.spotify.com](https://developer.spotify.com) → **Dashboard** → **Create app**
2. Fill in a name and description
3. Add `https://localhost:8888/callback` as a redirect URI
4. Note the **Client ID** and **Client Secret**

### 1b. Create the three playlists

Create three playlists in your Spotify account:

| Purpose | Env var |
|---|---|
| Main dynamic playlist (tracks live here) | `SPOTIFY_PLAYLIST_ID` |
| Genre rotation playlist | `GENRE_SPOTIFY_PLAYLIST_ID` |
| Top-50 high scores playlist | `HIGH_SCORES_SPOTIFY_PLAYLIST_ID` |

The playlist ID is the string after `/playlist/` in the Spotify share URL, e.g. `https://open.spotify.com/playlist/<ID>`.

### 1c. Obtain the OAuth token

`SPOTIFY_TOKEN` is a one-time setup using the auth tool in `scripts/spotify-auth-code-flow/`:

Spotify requires HTTPS for redirect URIs. Use [mkcert](https://github.com/FiloSottile/mkcert) to generate a locally-trusted certificate — this avoids browser security warnings without any manual cert trust steps.

Install mkcert (once per machine):

```sh
# macOS
brew install mkcert

# Linux
apt install mkcert        # Debian/Ubuntu
# or download the binary from https://github.com/FiloSottile/mkcert/releases
```

Generate the certificate (run from inside `scripts/spotify-auth-code-flow/`):

```sh
mkcert -install   # installs the local CA into your system trust store (once)
mkcert localhost  # generates localhost.pem and localhost-key.pem
```

Then start the server:

```sh
cd scripts/spotify-auth-code-flow
cp .env.example .env
# fill in SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET in .env
npm install
npm start
```

> **Codespaces note:** run this tool on your **local machine**, not inside a Codespace. `mkcert -install` adds the CA to your local system trust store — running it inside a container installs it there instead, so your browser still shows a cert warning. VS Code Desktop users can alternatively run the tool inside a Codespace (ports tunnel to localhost), but they must run `mkcert -install` locally first and copy the generated `.pem` files into the container.

Then visit `https://localhost:8888/login` in a browser. After authorizing, the callback returns a JSON response like:

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

Construct `SPOTIFY_TOKEN` from that response in the format `accessToken|refreshToken|tokenType|expiryUnix`, where `expiryUnix` is the current Unix timestamp plus `expires_in`:

```sh
# example (replace values from the JSON response)
echo "ACCESS_TOKEN|REFRESH_TOKEN|Bearer|$(( $(date +%s) + 3600 ))"
```

Store the resulting string as the `SPOTIFY_TOKEN` secret. Once stored, the embedded refresh token means the access token renews automatically on every run — the secret never needs to be updated again.

---

## 2. Neon

### 2a. Create a project

Sign in to [neon.tech](https://neon.tech) and create a new project. Choose a region close to your GitHub Actions runner region.

### 2b. Note the Project ID

The Project ID is shown in the project dashboard URL and in **Project Settings → General**. You will need it as `NEON_PROJECT_ID` for Codespaces.

### 2c. Get the main branch connection details

From **Project Dashboard → Connection Details**, select the `main` branch. You will need the host, port, user, password, and database name for the GitHub Actions secrets.

### 2d. Generate an API key

Go to your **Account Settings → API Keys → Generate new API key**. This is used by the devcontainer to create and destroy Neon branches. Store it as `NEON_API_KEY`.

---

## 3. GitHub Actions secrets

Go to **Repository → Settings → Secrets and variables → Actions → New repository secret**.

Add each of the following:

| Secret name | Value |
|---|---|
| `SPOTIFY_CLIENT_ID` | Spotify app client ID |
| `SPOTIFY_CLIENT_SECRET` | Spotify app client secret |
| `SPOTIFY_TOKEN` | Serialized OAuth token (see §1c) |
| `SPOTIFY_PLAYLIST_ID` | Main dynamic playlist ID |
| `GENRE_SPOTIFY_PLAYLIST_ID` | Genre rotation playlist ID |
| `HIGH_SCORES_SPOTIFY_PLAYLIST_ID` | Top-50 playlist ID |
| `NEON_HOST` | Neon main branch hostname |
| `NEON_PORT` | Neon port (usually `5432`) |
| `NEON_USER` | Neon role/user |
| `NEON_PASSWORD` | Neon password |
| `NEON_DB` | Neon database name |

> **Note:** The workflows map `NEON_*` secrets to `POSTGRES_*` environment variables internally. The secret names in GitHub must be the `NEON_*` names exactly as listed above.

---

## 4. Codespaces secrets

Codespaces secrets are separate from Actions secrets. They can be set at the account level (shared across repos) or at the repository level.

### Option A — Account level (recommended)

Go to [github.com/settings/codespaces](https://github.com/settings/codespaces) → **New secret**.

After creating each secret, use **Repository access** to grant it to this repository.

### Option B — Repository level

Go to **Repository → Settings → Secrets and variables → Codespaces → New repository secret**.

### Secrets to add

| Secret name | Value |
|---|---|
| `SPOTIFY_CLIENT_ID` | Spotify app client ID |
| `SPOTIFY_CLIENT_SECRET` | Spotify app client secret |
| `SPOTIFY_TOKEN` | Serialized OAuth token (see §1c) |
| `SPOTIFY_PLAYLIST_ID` | Main dynamic playlist ID |
| `GENRE_SPOTIFY_PLAYLIST_ID` | Genre rotation playlist ID |
| `HIGH_SCORES_SPOTIFY_PLAYLIST_ID` | Top-50 playlist ID |
| `NEON_API_KEY` | Neon API key (see §2d) |
| `NEON_PROJECT_ID` | Neon project ID (see §2b) |

> **Note:** `POSTGRES_*` variables are **not** needed as Codespaces secrets. The devcontainer startup script creates a Neon branch and writes them into `.env` automatically.

---

## 5. Opening a codespace

1. From the repository on GitHub, click **Code → Codespaces → New codespace**
2. Select the branch you want to work on
3. The container will start, install `neonctl`, and run `.devcontainer/scripts/neon-branch-setup.sh`
4. That script creates a Neon branch named `dev-<your-git-branch>` and populates `.env` with its connection details
5. If the branch was freshly created, run migrations before using the database:
   ```sh
   go run ./cmd/migration_runner
   ```

---

## 6. Cleaning up Neon branches

Neon branches are not deleted automatically. When you are finished with a feature branch, run:

```sh
bash .devcontainer/scripts/neon-branch-teardown.sh
```

This will prompt for confirmation and then delete the `dev-<your-git-branch>` Neon branch for the current git branch.
