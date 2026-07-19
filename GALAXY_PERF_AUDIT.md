# Playlist Galaxy — Performance Audit

Remaining work only. What's already shipped lives in the code and
`feat/galaxy-time-ranges`'s git history, not here.

## Open

- **`/api/graph` isn't actually edge-cached.** The `Cache-Control` header
  on a Cloudflare Pages Function response only governs the requesting
  browser's cache — Cloudflare doesn't edge-cache it just because the
  header is set. Every unique visitor still round-trips to Neon. Needs
  explicit caching via the Cache API (`caches.default.put`/`match`),
  keyed on the request URL, TTL matched to `refresh_graph_mv`'s 6h
  cadence.

- **`showArtists` toggle still triggers a full graph rebuild.** Unlike the
  time-window slider (which filters via a Sigma reducer with no rebuild),
  toggling "Show artists" in `galaxy/+page.svelte` still reruns
  `buildMixedGraph` from scratch — full node/edge rebuild, Louvain rerun,
  fresh FA2 layout. Same fix pattern as the window slider applies: make it
  a reducer-based show/hide instead of changing what's fed into
  `buildMixedGraph`.

- **`refresh_graph_mv` runtime at production scale is unconfirmed.** The
  gaps-and-islands ranges computation added in `migration013` was only
  timed against a scratch branch. `genre_graph_edges` is the priciest
  join in that migration — worth timing a refresh against real data
  volume before this lands anywhere that matters.
