// A closed interval [startEpochSeconds, endEpochSeconds] during which a
// node/edge had archive representation. Nodes/edges may carry several,
// non-overlapping ranges (e.g. a genre that fell off the playlist for a
// while and later came back).
export type TimeRange = [number, number];

// True if any of `ranges` overlaps the closed window [windowStart, windowEnd].
// Used to filter the mixed graph to a sliding time window entirely
// client-side — the server ships the full all-time graph plus these ranges
// and does no date filtering itself.
export function rangesOverlap(
	ranges: TimeRange[],
	windowStart: number,
	windowEnd: number,
): boolean {
	return ranges.some(
		([start, end]) => start <= windowEnd && end >= windowStart,
	);
}

// Formats a window's [start, end) epoch-second bounds as a short display
// range for the galaxy page's slider label, e.g. "Jun 1 – Jun 15, 2026".
export function formatWindowRange(
	windowStart: number,
	windowEnd: number,
): string {
	const day = new Intl.DateTimeFormat("en-US", {
		month: "short",
		day: "numeric",
	});
	const dayWithYear = new Intl.DateTimeFormat("en-US", {
		month: "short",
		day: "numeric",
		year: "numeric",
	});
	const start = new Date(windowStart * 1000);
	const end = new Date(windowEnd * 1000);
	return `${day.format(start)} – ${dayWithYear.format(end)}`;
}

// Formats the detail-drawer rank line, e.g. "Artist Rank: 1 out of 629".
// Rank is always all-time (see artist_rank/genre_rank materialized views) —
// it does not change when the galaxy page's time-window filter is active.
export function formatRank(
	kind: "artist" | "genre",
	rank: number,
	total: number,
): string {
	const label = kind === "artist" ? "Artist Rank" : "Genre Rank";
	return `${label}: ${rank.toLocaleString()} out of ${total.toLocaleString()}`;
}

// Log-scale node diameter
export function nodeSize(count: number, maxCount: number): number {
	return 14 + 50 * (Math.log(count + 1) / Math.log(maxCount + 1));
}

// Sqrt-scale edge width
export function edgeWidth(count: number, maxShared: number): number {
	return 0.5 + 5 * Math.sqrt(count / maxShared);
}

// Quartic-scale edge opacity, used to fade thin edges relative to the
// strongest edge of the same kind (edge kinds live on different weight
// scales, so each kind must be normalized against its own max, not a global
// one). Raising t to the 4th power pushes weak/mid-weight edges even closer
// to invisible than cubing did, and the ceiling is dropped hard (0.53 →
// 0.315 → 0.12): with thousands of overlapping edges, alpha compositing
// stacks even "low" per-edge opacity into a bright haze, so the ceiling has
// to be much lower than it would need to be for a sparse graph.
export function edgeOpacity(count: number, maxCount: number): number {
	const t = maxCount > 0 ? count / maxCount : 0;
	return 0.01 + 0.11 * t * t * t * t;
}

/**
 * Position for a degree-0 (isolated) node: its own community's cluster
 * center, or the graph centroid if it has no community. Keeping isolated
 * nodes anchored here (and marking them `fixed` for FA2) stops them from
 * drifting away from the cluster during force simulation, since they have
 * no edges to pull them back toward their community.
 */
export function isolatedNodePosition(
	community: number | undefined,
	communityCenters: Map<number, { x: number; y: number }>,
): { x: number; y: number } {
	if (community !== undefined) {
		const center = communityCenters.get(community);
		if (center) return center;
	}
	return { x: 0, y: 0 };
}

// 12 hues stepped by 150° so consecutive community IDs look maximally different.
// Hex values (not HSL) — sigma's WebGL renderer parses hex/rgb only.
export const COMMUNITY_PALETTE: readonly string[] = [
	"#DA654E", // hsl(10,  65%, 58%) coral
	"#4EDAAB", // hsl(160, 65%, 58%) mint
	"#DA4EC2", // hsl(310, 65%, 58%) magenta
	"#7DDA4E", // hsl(100, 65%, 58%) lime
	"#654EDA", // hsl(250, 65%, 58%) violet
	"#DAAB4E", // hsl(40,  65%, 58%) gold
	"#4EC2DA", // hsl(190, 65%, 58%) sky
	"#DA4E7D", // hsl(340, 65%, 58%) rose
	"#4EDA65", // hsl(130, 65%, 58%) emerald
	"#AB4EDA", // hsl(280, 65%, 58%) purple
	"#C2DA4E", // hsl(70,  65%, 58%) chartreuse
	"#4E7DDA", // hsl(220, 65%, 58%) cornflower
];

/**
 * Assigns each Louvain community a palette color so that same color always
 * means same community. Community IDs are sorted numerically and mapped to
 * palette entries in order; colors cycle if communities exceed palette length.
 *
 * Returns a Map<communityId, color>.
 */
export function assignCommunityColors(
	communityIds: Iterable<number>,
	palette: readonly string[],
): Map<number, string> {
	const sorted = [...new Set(communityIds)].sort((a, b) => a - b);
	const result = new Map<number, string>();
	sorted.forEach((commId, i) => {
		result.set(commId, palette[i % palette.length]);
	});
	return result;
}

/**
 * Blends a hex color toward its own perceived-luminance gray, used to give
 * artist nodes a visibly less saturated fill than genre nodes sharing the
 * same community color — same hue, muted intensity — rather than introducing
 * an unrelated color.
 *
 * `amount` is 0 (unchanged) to 1 (fully gray).
 */
export function desaturateColor(hex: string, amount: number): string {
	const r = Number.parseInt(hex.slice(1, 3), 16);
	const g = Number.parseInt(hex.slice(3, 5), 16);
	const b = Number.parseInt(hex.slice(5, 7), 16);
	const gray = 0.299 * r + 0.587 * g + 0.114 * b;
	const mix = (c: number) => Math.round(c + (gray - c) * amount);
	const toHex = (c: number) => c.toString(16).padStart(2, "0");
	return `#${toHex(mix(r))}${toHex(mix(g))}${toHex(mix(b))}`;
}

export interface SelectedNode {
	kind: "genre" | "artist";
	id: number;
}

export interface SearchableNode {
	id: number;
	kind: "genre" | "artist";
	name: string;
}

/**
 * Ranks nodes by name match against `query`: names that start with the query
 * rank above names that merely contain it, then shorter names, then
 * alphabetical. Purely client-side — the caller already has the full node
 * list loaded to build the graph, so this just filters/sorts it.
 */
export function searchNodes(
	query: string,
	nodes: SearchableNode[],
	limit = 8,
): SearchableNode[] {
	const q = query.trim().toLowerCase();
	if (!q) return [];

	const matches: SearchableNode[] = [];
	for (const node of nodes) {
		if (node.name.toLowerCase().includes(q)) matches.push(node);
	}

	matches.sort((a, b) => {
		const aStarts = a.name.toLowerCase().startsWith(q) ? 0 : 1;
		const bStarts = b.name.toLowerCase().startsWith(q) ? 0 : 1;
		if (aStarts !== bStarts) return aStarts - bStarts;
		if (a.name.length !== b.name.length) return a.name.length - b.name.length;
		return a.name.localeCompare(b.name);
	});

	return matches.slice(0, limit);
}

// Parses the `?node=g:14` / `?node=a:7` deep-link query param used by the
// graph page's detail drawer. Returns null for anything malformed so callers
// never need to special-case a bad/stale link.
export function parseNodeParam(value: string | null): SelectedNode | null {
	if (!value) return null;
	const match = /^(g|a):(\d+)$/.exec(value);
	if (!match) return null;
	const id = Number(match[2]);
	if (!Number.isInteger(id) || id <= 0) return null;
	return { kind: match[1] === "g" ? "genre" : "artist", id };
}

// Parses the `?start=&end=` epoch-second window shared by /api/genres/:id
// and /api/artists/:id. Both params must be present and form a valid
// [start, end] pair — anything else (missing, non-numeric, inverted) is
// treated as "no time filter" rather than a request error, since a stale or
// malformed link should fall back to all-time data instead of failing.
export function parseTimeRangeParams(
	searchParams: URLSearchParams,
): TimeRange | null {
	const startParam = searchParams.get("start");
	const endParam = searchParams.get("end");
	if (startParam === null || endParam === null) return null;
	const start = Number(startParam);
	const end = Number(endParam);
	if (!Number.isFinite(start) || !Number.isFinite(end) || start > end) {
		return null;
	}
	return [start, end];
}

// Builds the detail-drawer fetch URL for a selected node, appending the
// active time window as `start`/`end` query params when one is active.
export function detailFetchUrl(
	node: SelectedNode,
	activeWindow: TimeRange | null,
): string {
	const base =
		node.kind === "genre"
			? `/api/genres/${node.id}`
			: `/api/artists/${node.id}`;
	if (!activeWindow) return base;
	return `${base}?start=${activeWindow[0]}&end=${activeWindow[1]}`;
}

// Converts an inclusive [start, end] window into the half-open bounds a SQL
// `created_at` filter needs. `end` names a whole second, but it was derived
// by flooring a (sub-second-precision) timestamp — see the graph_vertices/
// edges materialized views — so the row it's meant to include usually has a
// fractional second past it. Filtering with `created_at < endExclusive`
// against the bare column (rather than flooring `created_at` itself before
// comparing) keeps the comparison sargable against idx_song_archive_created_at.
export function sqlBounds(range: TimeRange): {
	start: number;
	endExclusive: number;
} {
	return { start: range[0], endExclusive: range[1] + 1 };
}
