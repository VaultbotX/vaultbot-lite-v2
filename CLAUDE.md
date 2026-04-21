# Vaultbot Lite — Claude Guide

## What this is

A Spotify playlist tracker that polls a playlist on a schedule, stores tracks and their metadata (artists, genres) in a Neon PostgreSQL database, and exposes the data via a SvelteKit web app deployed to Cloudflare Pages. Go handles all data collection and curation; the frontend is read-only.

## Hard constraints

- **No direct DB access from the frontend.** All database queries live in SvelteKit API routes (`+server.ts`) which run as Cloudflare Pages Functions. Page components only call those routes via `fetch`.
- **No SSR page rendering depends on secrets.** Pages use universal load functions (`+page.ts`, not `+page.server.ts`) that call the API routes. Keep secrets in the server-side `+server.ts` routes only.
- **No Tailwind.** Styles live in scoped `<style>` blocks inside `.svelte` files, or in `src/app.css` for globals.
- **Migrations are Go structs, not SQL files.** Never create raw `.sql` migration files; follow the pattern in `internal/persistence/postgres/migrations/`.
- **Workflow files cannot be pushed by Claude.** The GitHub OAuth token used in this environment does not have the `workflow` scope. If a workflow needs to be added or changed, paste the YAML in chat and the human commits it manually. See `LEARNINGS.md`.

## Tech stack

| Concern | Tool |
|---|---|
| Data collection | Go 1.26, scheduled via GitHub Actions |
| Database | Neon PostgreSQL (serverless) |
| Frontend framework | SvelteKit 2 + Svelte 5 |
| Frontend adapter | `@sveltejs/adapter-cloudflare` |
| Frontend deploy | Cloudflare Pages |
| DB driver (frontend) | `@neondatabase/serverless` (HTTP mode) |
| Linting/formatting | Biome 2.x (covers `.ts` and `.svelte`) |
| Graph visualization | Cytoscape.js (with Louvain community detection) |
| Charts/treemap | Chart.js 4 + chartjs-chart-treemap |
| Unit testing | Vitest (pure function tests in `src/**/*.test.ts`) |

## Environment setup

Run `.devcontainer/scripts/setup.sh` from the repo root to scaffold a fresh environment:

```sh
bash .devcontainer/scripts/setup.sh
```

This does, in order:

1. `npm install` inside `web/` — installs all frontend dependencies
2. `cp web/.dev.vars.example web/.dev.vars` — creates a local env file for `wrangler pages dev` (skipped if already present)
3. `npx svelte-kit sync` inside `web/` — generates `.svelte-kit/tsconfig.json` and `$types` imports that `tsconfig.json` extends

**Always run `setup.sh` (or its steps manually) before building, type-checking, or running the frontend.** Without step 3, TypeScript fails with `tsconfig not found` errors because `web/tsconfig.json` extends `.svelte-kit/tsconfig.json`.

In a devcontainer, this runs automatically as part of `postCreateCommand`.

### Frontend environment variables

`web/.dev.vars` holds secrets for local `wrangler pages dev`. It is gitignored. The only required variable is:

| Variable | Description |
|---|---|
| `DATABASE_URL` | Neon connection string (`postgresql://user:pass@host/db?sslmode=require`) |

In production, set this in the Cloudflare Pages dashboard under **Settings → Environment variables**.

## Key commands

### Frontend (`cd web` first)

```sh
npm run dev          # Vite dev server (no Cloudflare bindings)
npm run build        # Production build via adapter-cloudflare
npm run check        # svelte-kit sync + svelte-check type checking
npm run test         # Run unit tests (Vitest)
npm run test:watch   # Run unit tests in watch mode
npm run biome        # Biome lint + format (auto-fix)
npm run lint         # Biome lint only
npm run format       # Biome format only
```

### Go (from repo root)

```sh
go build ./...                          # Build all binaries
go test ./...                           # Run all tests
go run ./cmd/migration_runner           # Apply pending DB migrations
go run ./cmd/refresh_graph_mv           # Refresh genre graph materialized views
go run ./cmd/poll                       # Poll Spotify playlist once
go run ./cmd/stats                      # Generate stats JSON (stdout) — superseded by /api/stats
```

## Project structure

```
.
├── cmd/                        # Go executable entry points
│   ├── migration_runner/       # Applies DB migrations
│   ├── refresh_graph_mv/       # Refreshes genre graph MVs
│   ├── poll/                   # Spotify polling job
│   ├── stats/                  # Legacy stats JSON generator (superseded by /api/stats)
│   ├── purge/                  # Removes expired tracks
│   ├── genre/                  # Genre rotation playlist
│   ├── highscores/             # Top-50 playlist
│   ├── throwback/              # Throwback playlist
│   └── variety/                # Variety playlist
├── internal/
│   ├── persistence/postgres/
│   │   ├── migrations/         # Migration definitions (Go structs with SQL)
│   │   ├── archive/            # song_archive queries
│   │   ├── artists/            # artists queries
│   │   ├── genres/             # genres queries
│   │   └── songs/              # songs queries
│   ├── cron/                   # Playlist curation logic
│   ├── domain/                 # Business logic and interfaces
│   ├── spotify/                # Spotify API client
│   └── utils/
├── web/                        # SvelteKit frontend
│   ├── src/
│   │   ├── app.css             # Global styles + CSS variables
│   │   ├── app.d.ts            # App.Platform type (Cloudflare env bindings)
│   │   ├── routes/
│   │   │   ├── +layout.svelte  # Root layout (header, nav, footer)
│   │   │   ├── +page.svelte    # Stats dashboard (summary cards + charts)
│   │   │   ├── +page.ts        # Fetches /api/stats
│   │   │   ├── graph/          # Interactive genre graph page
│   │   │   │   ├── +page.svelte
│   │   │   │   └── +page.ts    # Fetches /api/graph
│   │   │   ├── genre/[id]/     # Genre drilldown page
│   │   │   └── api/            # Server-side API routes (Cloudflare Pages Functions)
│   │   │       ├── stats/      # GET /api/stats
│   │   │       ├── graph/      # GET /api/graph
│   │   │       └── genre/[id]/ # GET /api/genre/:id
│   │   └── lib/
│   │       ├── GenreGraph.svelte   # Cytoscape.js graph component
│   │       ├── StatsCharts.svelte  # Chart.js charts component (line, bar, treemap)
│   │       ├── allNamed.ts         # Parallel DB query helper
│   │       ├── graph.ts            # Pure fns: communityColor, nodeSize, edgeWidth
│   │       ├── graph.test.ts       # Unit tests for graph.ts
│   │       ├── louvain.ts          # Louvain community detection algorithm
│   │       ├── louvain.test.ts     # Unit tests for louvain.ts
│   │       ├── stats.ts            # Pure fns: fmtMonth, treemapColor
│   │       └── stats.test.ts       # Unit tests for stats.ts
│   ├── static/                 # Static assets (logo, favicon)
│   ├── biome.json              # Biome linter config
│   ├── vitest.config.ts        # Vitest config (separate from vite.config.ts)
│   └── wrangler.toml           # Cloudflare Pages config (update `name`)
└── .devcontainer/
    └── scripts/
        ├── setup.sh            # Frontend bootstrap (npm install, svelte-kit sync)
        ├── neon-branch-setup.sh
        └── neon-branch-teardown.sh
```

## Database schema (key tables)

```
songs          ←→ link_song_artists  ←→ artists
                                            ↕
songs          ←→ link_song_genres   ←→ genres
                                            ↕
artists        ←→ link_artist_genres ←→ genres

songs          ←→ song_archive       (timestamped occurrence log)

-- Materialized views (updated every 6 hours via refresh_graph_mv)
genre_graph_vertices   (genre_id, name, artist_count)
genre_graph_edges      (source_genre_id, target_genre_id, shared_artist_count)
```

## Adding a DB migration

1. Create `internal/persistence/postgres/migrations/migration0NN.go`:

```go
package migrations

var Migration0NN = &Migration{
    Name: "0NN-DescriptiveName",
    Up:   `ALTER TABLE ...`,
    Down: ``,
}
```

2. Register it in `cmd/migration_runner/runner.go` — increment the array size and append the variable.

## Testing

Unit tests live alongside their source files in `web/src/lib/` as `*.test.ts`. They cover pure functions only — no Svelte components, no DOM, no DB.

```
web/src/lib/
├── graph.test.ts       # communityColor, nodeSize, edgeWidth
├── louvain.test.ts     # detectCommunities (Louvain algorithm)
└── stats.test.ts       # fmtMonth, treemapColor
```

**Why a separate `vitest.config.ts`:** Vite 8 uses the rolldown backend whose `Plugin` type is incompatible with Vitest's bundled Vite version. Importing `defineConfig` from `vitest/config` inside `vite.config.ts` causes a type conflict. The fix is a standalone `vitest.config.ts` that imports from `vitest/config`, leaving `vite.config.ts` untouched.

**Why `npm run check` must run before `npm test` in CI:** Vitest's esbuild resolves `web/tsconfig.json → .svelte-kit/tsconfig.json`, which only exists after `svelte-kit sync` runs. `npm run check` internally runs `svelte-kit sync`, so order the CI steps: Lint → Type check (`npm run check`) → Test (`npm test`).

**Pure function extraction pattern:** Move any logic that doesn't depend on DOM, Svelte reactivity, or Chart.js/Cytoscape instances into a plain `.ts` file in `lib/`. This makes it directly testable with Vitest. Keep chart and graph configuration inside the component's `onMount` / `$effect`.

## TypeScript conventions

**Use types as strictly as possible. Avoid `any` and avoid `as unknown as T` unless there is no other option.**

### Parallel DB queries — always use `allNamed` + `typed`

`Promise.all` with positional destructuring is fragile: inserting or reordering a query silently mismatches results to variable names with no compile-time error. Always use the helpers in `web/src/lib/allNamed.ts` instead:

```typescript
import { allNamed, typed } from "$lib/allNamed";

const { genreRows, artists } = await allNamed({
  genreRows: typed<{ name: string }[]>(sql`SELECT name FROM genres WHERE id = ${id}`),
  artists:   typed<Artist[]>(sql`SELECT name FROM artists WHERE ...`),
});
```

- **`allNamed`** — runs promises in parallel keyed by name; position in the object is irrelevant.
- **`typed<T>`** — the single approved boundary cast for neon query results. The neon template tag returns `NeonQueryPromise<…, Record<string,any>[]>` which TypeScript 6 will not let you narrow with a plain `as`. `typed<T>` confines the required `as unknown as` to one documented location.

Never scatter `as unknown as SomeType` across route files. If a new external boundary needs a cast, add a named helper like `typed` rather than inlining it.

### `satisfies` for return shapes

Use `satisfies` (not `as`) when asserting that a value matches an interface at a return site — it checks the shape without widening the type:

```typescript
return json({ vertices, edges } satisfies GraphData, { headers: { … } });
```

## Svelte 5 patterns

Use runes throughout. No `$:` reactive declarations, no `export let`.

```svelte
<script lang="ts">
  let { value = "default" }: { value?: string } = $props();
  let count = $state(0);
  let doubled = $derived(count * 2);
</script>
```

Use `onclick`, `onchange` (not `on:click`) for DOM element events.

## Design tokens

The entire UI uses these CSS variables (defined in `web/src/app.css`):

| Variable | Value | Use |
|---|---|---|
| `--bg` | `#0c0c10` | Page background |
| `--surface` | `#131318` | Cards, header |
| `--surface-2` | `#1a1a22` | Nested surfaces |
| `--border` | `#1f1f2c` | Borders, dividers |
| `--text` | `#e2e2f0` | Primary text |
| `--text-muted` | `#6060a0` | Secondary text |
| `--accent` | `#7c6af7` | Purple accent, links |
| `--radius` | `10px` | Border radius |

Fonts: **IBM Plex Sans** (body, 400/500/600) and **IBM Plex Mono** (numbers, metadata, code).
