#!/usr/bin/env bash
set -euo pipefail

# Validate required secrets are present
: "${NEON_API_KEY:?NEON_API_KEY secret is required}"
: "${NEON_PROJECT_ID:?NEON_PROJECT_ID secret is required}"

# Read the branch name written by neon-branch-setup.sh — this is the
# authoritative name actually used when the branch was created, which
# may differ from what we'd derive from the current git branch name.
NEON_BRANCH_NAME=$(grep "^NEON_BRANCH_NAME=" .env 2>/dev/null | cut -d'=' -f2-)

if [[ -z "${NEON_BRANCH_NAME}" ]]; then
  echo "NEON_BRANCH_NAME not found in .env — was neon-branch-setup.sh run in this environment?" >&2
  exit 1
fi

echo "Neon branch: ${NEON_BRANCH_NAME}"
echo ""

# Confirm before deleting
read -r -p "Delete Neon branch '${NEON_BRANCH_NAME}'? [y/N] " confirm
if [[ "${confirm}" != [yY] ]]; then
  echo "Aborted."
  exit 0
fi

echo "Deleting Neon branch '${NEON_BRANCH_NAME}'..."
neonctl branch delete "${NEON_BRANCH_NAME}" \
  --project-id "${NEON_PROJECT_ID}" \
  --confirm

echo "Branch '${NEON_BRANCH_NAME}' deleted."
