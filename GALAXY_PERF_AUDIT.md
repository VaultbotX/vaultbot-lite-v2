# Playlist Galaxy — Performance Audit (living doc)

Scope: three issues pulled out of the broader galaxy perf pass for focused
fix planning. Updated as we investigate/fix each one.

---

## 1. Dynamic query path (`/api/graph?dynamic=true`)

**File:** `web/src/routes/api/graph/+server.ts`

**The gap:** the dynamic branch issues 5 queries via `allNamed`, and 4 of
them (`genreVertices`, `artistVertices`, `genreGenreEdges`,
`genreArtistEdges`) independently redeclare and re-execute the same two
CTEs — `CANONICAL_ARCHIVE_CTE` (full `song_archive` ⋈ `songs` ⋈
`duplicate_song_lookup`, aggregated) and `recent_artist_ids` /
`artist_archive`. The neon HTTP driver has no implicit batching across
separate `sql\`...\`` calls, so each of those 4 is a fully separate round
trip to Neon that redoes the same expensive join+aggregate from scratch.
That's ~4x the necessary DB work per request, on the query path that's
already the "expensive" one relative to the materialized-view-backed
static path.

**Second, separate gap — the recency filter is a membership filter, not a
value filter.** `recent_artist_ids` restricts *which* artists/genres show
up (must have a `song_archive` row in the last 14 days), but the weights
attached to those nodes/edges (`archive_count`, `shared_archive_count`,
`shared_song_count`) come from `canonical_archive` / `v_songs` joins with
**no date bound** — i.e. they're all-time totals. So "dynamic" mode today
means "the all-time graph, filtered down to nodes that have been touched
recently," not "the graph as it looked over the last 14 days." Node sizes
and edge weights don't actually reflect recent activity, only recent
*presence*. This matters directly for the sliding-window discussion below
— it's not a small tweak away from a real window, the weighting model
would need to change too.

**Third:** no edge caching (see #2) means this redundant work runs on
every cache-miss request, not just occasionally.

---

## 2. Cache-Control on Cloudflare Pages Functions

**File:** `web/src/routes/api/graph/+server.ts:218-220`

**The gap:** the route sets `Cache-Control: public, max-age=...` (21600s
static / 300s dynamic) but this is a Cloudflare Pages *Function* response
— it's dynamic compute, not a static asset, so Cloudflare's edge does not
cache it automatically just because the header is present. The header only
governs the *requesting browser's* HTTP cache. Net effect: every unique
visitor (or any visitor past their own browser's cache window) round-trips
to Neon, for both the static graph (which only changes every 6h, on
`refresh_graph_mv`'s schedule) and the dynamic graph (which is the
expensive query from #1). We're not actually getting the caching behavior
the header implies. Fix direction: explicit edge caching via the Cache API
(`caches.default.put`/`match`) keyed on the request URL, with TTLs matched
to the existing max-age values (or to the materialized-view refresh
cadence for the static path).

---

## 3. Client-side graph (re)creation (`GenreGraph.svelte`, `mixed-graph.ts`)

**Files:** `web/src/routes/galaxy/+page.svelte`,
`web/src/lib/mixed-graph.ts`, `web/src/lib/GenreGraph.svelte`

**The gap:** `buildMixedGraph` is a `$derived` keyed on `showArtists` and
`showDynamic`/`dynamicData`. Toggling either checkbox rebuilds the entire
graphology graph from scratch — every node/edge re-added, Louvain
community detection re-run, all node sizes/colors recomputed — which in
turn tears down and recreates the whole Sigma instance and reruns
`initCommunityLayout` + 500 iterations of ForceAtlas2. So "hide artist
nodes," a pure visibility change, currently costs a full physics
relayout. FA2 already runs on the main thread (not a worker) — there's a
comment noting Barnes-Hut was added specifically because the O(n²) exact
version was blocking the main thread for seconds at ~1,300 nodes — so any
trigger that reruns it is a real, felt UI freeze, not just wasted CPU.

This is the one most directly in tension with the sliding-window idea
below: a slider that changes the dataset on every drag tick would hit this
same full-rebuild path on every tick unless it's fixed first.

---

## Decision: kill the dynamic route, ship time-range metadata, filter entirely client-side

Agreed direction: `/api/graph?dynamic=true` goes away completely. There's
one graph response, always the full all-time node/edge set (same as
today's static path), and every node and edge additionally carries
metadata describing *when* it had representation. The server does zero
date filtering — filtering, bucketing-into-a-window, and rendering
decisions are entirely a client concern. This resolves #1 and #2 outright
(one query shape, one cache tier) and turns out to also be the cleanest
fix for #3. Details below.

### What "time range of representation" actually means here

There's no single clean timestamp signal to draw on — weight today is
already a rollup (see `migration012.go`):

- An **artist's** archive activity = `song_archive.created_at` values for
  that artist's canonical songs (via `v_songs` + the duplicate-lookup
  join).
- A **genre's** archive activity = the *union* of its artists' archive
  activity (via `link_artist_genres`) — a genre has no direct timestamp of
  its own, it inherits from whichever artists carry it.
- An **artist-artist edge's** activity = timestamps from the *shared*
  songs specifically, not just "both artists were active around the same
  time."
- A **genre-genre edge's** activity = timestamps from the shared artist(s)
  that carry both genres.
- A **genre-artist edge** = the artist's own activity, scoped to songs
  that also carry that genre.

This matters because edges need **their own ranges**, computed from the
shared/overlapping activity specifically — deriving an edge's visibility
from "both endpoint nodes happen to have an overlapping range" would show
false collabs (two artists each active in the window, but not *together*
in it).

### Proposed representation: collapsed intervals, not raw timestamps or day-buckets

Rather than shipping every `song_archive.created_at` (unbounded, grows
forever) or fixed day-buckets (arbitrary resolution, still a lot of
zero-filled rows for sparse nodes), compute **gaps-and-islands** per node
and per edge: collapse the sorted timestamps into `[start, end]` runs,
treating a gap larger than some threshold (recommend ~18–24h — 3–4 missed
6-hourly polls) as a real absence rather than a skipped poll. Most nodes
that are continuously popular collapse to a single range; only nodes with
genuinely sporadic on/off history produce multiple ranges. This is a
standard SQL window-function pattern (`LAG`/`LEAD` over `created_at`,
partitioned per node, flagging new-island starts where the gap exceeds
the threshold), so it's compact for the common case and still exact.

Client-side, checking "is this node/edge active in the selected window"
is then just an O(ranges) overlap test — trivial even for thousands of
nodes on every slider tick.

### Where to compute it

The existing `refresh_graph_mv` cron already recomputes
`genre_graph_vertices`/`edges` and `artist_graph_vertices`/`edges` every
6h from these same joins. Extending those materialized views with a
`ranges` column (e.g. `jsonb`, array of `[start_epoch, end_epoch]` pairs)
is the natural home — it keeps "no filtering at request time" literally
true (the route becomes a plain `SELECT ... FROM genre_graph_vertices`,
same shape it already has for the static path today) and keeps this
computation off the request path entirely, on the same 6h cadence as
everything else. No live aggregation, no separate cache tier from #2 to
manage — everything lands on the existing 6h TTL.

### This is also the real fix for #3, not just an unrelated nice-to-have

Because the server always returns the *same* full graph regardless of the
selected window, the graph topology FA2/Louvain operate on never changes
when the window changes — only which nodes/edges are *visible* does.
`GenreGraph.svelte` already has exactly the mechanism for this: the
`nodeReducer`/`edgeReducer` pair currently used to dim non-neighbors on
hover/selection. Extending that same reducer to also dim/hide
out-of-window nodes and edges means changing the window becomes a redraw,
not a rebuild — no graph reconstruction, no Louvain rerun, no FA2 rerun.
That's what actually makes a slider (vs. a two-state toggle) viable; doing
the range-overlap filtering client-side is what unlocks that path, since
the server no longer needs to be involved at all once the metadata is on
the wire.

(Note `showArtists` today has the same rebuild-on-toggle problem and isn't
in scope here, but it's the same fix pattern — worth a follow-up once this
lands.)

### Open questions

- **Gap threshold** — is ~18–24h the right cutoff for "still the same
  run" vs. "ended and came back," or should it be tunable/derived from
  actual poll cadence rather than hardcoded?
- **Edge range computation cost** — genre-genre and artist-artist edges
  are already the most expensive joins in the current queries (self-joins
  over `link_artist_genres`/`link_song_artists`); adding a gaps-and-islands
  window function on top of those needs a sanity check against
  `refresh_graph_mv`'s actual runtime budget, even though it's off the
  request path.
- **Payload size** — should be modest (most nodes ≈ 1 range), but worth
  confirming against real data volume before committing, especially for
  high-churn genres/artists that could fragment into many short ranges.
- **Range precision** — epoch seconds vs. ISO strings vs. something more
  compact; minor, but worth picking once rather than per-endpoint.
