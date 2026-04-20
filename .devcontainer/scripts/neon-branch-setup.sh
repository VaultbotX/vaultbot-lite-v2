#!/usr/bin/env bash
set -euo pipefail

# Validate required secrets are present
: "${NEON_API_KEY:?NEON_API_KEY secret is required}"
: "${NEON_PROJECT_ID:?NEON_PROJECT_ID secret is required}"

# Derive a Neon-safe branch name from the current git branch.
# Neon branch names support alphanumeric, hyphens, underscores, and dots.
# Forward slashes (common in git branch names) are replaced with hyphens.
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
NEON_BRANCH_NAME="dev-$(printf '%s' "${GIT_BRANCH}" | tr '/' '-')"

echo "Git branch:  ${GIT_BRANCH}"
echo "Neon branch: ${NEON_BRANCH_NAME}"
echo ""

# Try to fetch the connection string first — if the branch already exists
# (e.g. container rebuild) this short-circuits without creating a duplicate.
echo "Checking for existing Neon branch..."
if ! CONN_URI=$(neonctl connection-string \
  --project-id "${NEON_PROJECT_ID}" \
  --branch "${NEON_BRANCH_NAME}" 2>/dev/null); then

  echo "Branch not found — creating '${NEON_BRANCH_NAME}'..."
  neonctl branch create \
    --project-id "${NEON_PROJECT_ID}" \
    --name "${NEON_BRANCH_NAME}"

  CONN_URI=$(neonctl connection-string \
    --project-id "${NEON_PROJECT_ID}" \
    --branch "${NEON_BRANCH_NAME}")
else
  echo "Branch already exists — skipping creation."
fi

echo "Connection URI retrieved."
echo ""

# URL-decode a percent-encoded string using only printf and bash.
# e.g. "p%40ssword" -> "p@ssword"
urldecode() {
  printf '%b' "${1//%/\\x}"
}

# Parse the connection URI with pure bash string operations.
# Format: postgresql://user:password@host:port/dbname?query
_rest="${CONN_URI#postgresql://}"
_userinfo="${_rest%%@*}"
_hostpath="${_rest#*@}"

PG_USER="${_userinfo%%:*}"
PG_PASSWORD="$(urldecode "${_userinfo#*:}")"

_hostport="${_hostpath%%/*}"
_pathquery="${_hostpath#*/}"

PG_HOST="${_hostport%%:*}"
if [[ "${_hostport}" == *:* ]]; then
  PG_PORT="${_hostport##*:}"
else
  PG_PORT="5432"
fi

PG_DB="${_pathquery%%\?*}"

# Upsert a KEY=VALUE line in a target file using only bash + standard POSIX utilities.
# Replaces the existing line for KEY if present; appends otherwise.
# Usage: set_var <file> <key> <value>
set_var() {
  local file="$1" key="$2" value="$3" tmp
  tmp="$(mktemp)"
  if grep -q "^${key}=" "${file}" 2>/dev/null; then
    while IFS= read -r line || [[ -n "${line}" ]]; do
      if [[ "${line}" == "${key}="* ]]; then
        printf '%s=%s\n' "${key}" "${value}"
      else
        printf '%s\n' "${line}"
      fi
    done < "${file}" > "${tmp}"
    mv "${tmp}" "${file}"
  else
    rm -f "${tmp}"
    printf '%s=%s\n' "${key}" "${value}" >> "${file}"
  fi
}

# Seed .env from .env.example if it doesn't exist yet
if [[ ! -f .env ]]; then
  cp .env.example .env
  echo "Created .env from .env.example"
fi

set_var .env "NEON_BRANCH_NAME"  "${NEON_BRANCH_NAME}"
set_var .env "POSTGRES_HOST"     "${PG_HOST}"
set_var .env "POSTGRES_PORT"     "${PG_PORT}"
set_var .env "POSTGRES_USER"     "${PG_USER}"
set_var .env "POSTGRES_PASSWORD" "${PG_PASSWORD}"
set_var .env "POSTGRES_DB"       "${PG_DB}"

echo "Postgres connection vars written to .env"
echo ""

# Write DATABASE_URL to web/.dev.vars for the SvelteKit frontend.
# setup.sh (which runs after this script) will skip its own bootstrap
# copy when it sees the file already exists.
if [[ ! -f web/.dev.vars ]]; then
  cp web/.dev.vars.example web/.dev.vars
  echo "Created web/.dev.vars from web/.dev.vars.example"
fi

set_var web/.dev.vars "DATABASE_URL" "${CONN_URI}"

echo "DATABASE_URL written to web/.dev.vars"
echo ""
echo "Done! Neon branch '${NEON_BRANCH_NAME}' is ready."
echo "If this is a freshly created branch, run migrations:"
echo "  go run ./cmd/migration_runner"
