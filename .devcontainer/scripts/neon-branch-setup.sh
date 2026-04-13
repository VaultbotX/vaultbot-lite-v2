#!/usr/bin/env bash
set -euo pipefail

# Validate required secrets are present
: "${NEON_API_KEY:?NEON_API_KEY secret is required}"
: "${NEON_PROJECT_ID:?NEON_PROJECT_ID secret is required}"

# Derive a Neon-safe branch name from the current git branch.
# Neon branch names support alphanumeric, hyphens, underscores, and dots.
# Forward slashes (common in git branch names) are replaced with hyphens.
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
NEON_BRANCH_NAME="dev-$(echo "${GIT_BRANCH}" | tr '/' '-')"

echo "Git branch:   ${GIT_BRANCH}"
echo "Neon branch:  ${NEON_BRANCH_NAME}"
echo ""

# Create the branch if it does not already exist.
# neonctl exits non-zero if the branch name conflicts, so we check first.
EXISTING=$(neonctl branch list --project-id "${NEON_PROJECT_ID}" --output json \
  | python3 -c "
import json, sys, os
branches = json.load(sys.stdin)
name = os.environ['NEON_BRANCH_NAME']
print('true' if any(b['name'] == name for b in branches) else 'false')
")

if [ "${EXISTING}" = "true" ]; then
  echo "Neon branch '${NEON_BRANCH_NAME}' already exists — skipping creation."
else
  echo "Creating Neon branch '${NEON_BRANCH_NAME}'..."
  neonctl branch create \
    --project-id "${NEON_PROJECT_ID}" \
    --name "${NEON_BRANCH_NAME}"
  echo "Branch created."
fi

echo ""
echo "Fetching connection details..."

# Retrieve the connection URI for the branch.
# neonctl handles waiting for the endpoint to become ready.
CONN_URI=$(neonctl connection-string \
  --project-id "${NEON_PROJECT_ID}" \
  --branch "${NEON_BRANCH_NAME}")

echo "Connection URI retrieved."
echo ""

# Parse the connection URI and update .env, using Python to avoid
# shell quoting issues with passwords containing special characters.
export CONN_URI

python3 - <<'PYEOF'
import os
import re
from urllib.parse import urlparse

conn_uri = os.environ["CONN_URI"]
p = urlparse(conn_uri)

updates = {
    "POSTGRES_HOST": p.hostname,
    "POSTGRES_PORT": str(p.port or 5432),
    "POSTGRES_USER": p.username,
    "POSTGRES_PASSWORD": p.password,
    "POSTGRES_DB": p.path.lstrip("/").split("?")[0],
}

env_path = ".env"

# Seed .env from .env.example if it doesn't exist yet
if not os.path.exists(env_path):
    if os.path.exists(".env.example"):
        with open(".env.example") as f:
            content = f.read()
        print(f"Created .env from .env.example")
    else:
        content = ""
else:
    with open(env_path) as f:
        content = f.read()

# Update existing keys or append missing ones
for key, value in updates.items():
    pattern = rf"^{key}=.*$"
    replacement = f"{key}={value}"
    if re.search(pattern, content, re.MULTILINE):
        content = re.sub(pattern, replacement, content, flags=re.MULTILINE)
    else:
        content = content.rstrip("\n") + f"\n{key}={value}\n"

with open(env_path, "w") as f:
    f.write(content)

print("Postgres connection vars written to .env")
PYEOF

echo ""
echo "Done! Neon branch '${NEON_BRANCH_NAME}' is ready."
echo "Run migrations if this is a freshly created branch:"
echo "  go run ./cmd/migration_runner"
