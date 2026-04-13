#!/usr/bin/env bash
set -euo pipefail

# Validate required secrets are present
: "${NEON_API_KEY:?NEON_API_KEY secret is required}"
: "${NEON_PROJECT_ID:?NEON_PROJECT_ID secret is required}"

GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
NEON_BRANCH_NAME="dev-$(echo "${GIT_BRANCH}" | tr '/' '-')"

echo "Git branch:   ${GIT_BRANCH}"
echo "Neon branch:  ${NEON_BRANCH_NAME}"
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
