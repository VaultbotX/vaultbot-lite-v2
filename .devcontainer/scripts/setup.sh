#!/usr/bin/env bash
set -e

echo "==> Setting up vaultbot-web frontend..."

cd web

# Install dependencies
npm install

# Bootstrap .dev.vars for local Cloudflare Pages dev if not already present.
# Real DATABASE_URL can be filled in later; the file must exist for wrangler pages dev.
if [ ! -f .dev.vars ]; then
	cp .dev.vars.example .dev.vars
	echo "Created web/.dev.vars from web/.dev.vars.example"
fi

# Generate SvelteKit type declarations ($app types, $types imports, path aliases, etc.)
# tsconfig.json extends .svelte-kit/tsconfig.json which is produced by this step.
# Without it, TypeScript and editor tooling fail with "tsconfig not found" errors.
npx svelte-kit sync

echo "==> Frontend setup complete."
