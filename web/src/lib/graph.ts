// A closed interval [startEpochSeconds, endEpochSeconds] during which a
// node/edge had archive representation. Nodes/edges may carry several,
// non-overlapping ranges (e.g. a genre that fell off the playlist for a
// while and later came back).
export type TimeRange = [number, number];

// True if any of `ranges` overlaps the closed window [windowStart, windowEnd].
// Used to filter the mixed graph to a sliding time window entirely
// client-side — the server ships the full all-time graph plus artist/collab
// ranges and does no date filtering itself.
export function rangesOverlap(
	ranges: TimeRange[],
	windowStart: number,
	windowEnd: number,
): boolean {
	return ranges.some(
		([start, end]) => start <= windowEnd && end >= windowStart,
	);
}

// Computes the overall [minStart, maxEnd] bounds across many range arrays
// iteratively (avoiding `.flatMap()` + `Math.max(...spread)`, which overflows
// the JS call stack on large archives). Returns null if no intervals exist.
export function computeTimeBounds(
	rangeLists: Iterable<TimeRange[]>,
): TimeRange | null {
	let min = Number.POSITIVE_INFINITY;
	let max = Number.NEGATIVE_INFINITY;
	for (const ranges of rangeLists) {
		for (const [start, end] of ranges) {
			if (start < min) min = start;
			if (end > max) max = end;
		}
	}
	if (!Number.isFinite(min) || !Number.isFinite(max)) return null;
	return [min, max];
}

export function edgePairKey(idA: number, idB: number): string {
	return idA < idB ? `${idA}:${idB}` : `${idB}:${idA}`;
}

export interface WindowIndex {
	artists: { artist_id: number; ranges: TimeRange[] }[];
	genreArtists: Map<number, number[]>;
	genreGenreSharedArtists: {
		pairKey: string;
		sourceGenreId: number;
		targetGenreId: number;
		sharedArtistIds: number[];
	}[];
	artistArtistEdges: {
		pairKey: string;
		sourceArtistId: number;
		targetArtistId: number;
		ranges: TimeRange[];
	}[];
}

export interface WindowVisibility {
	genreIds: Set<number>;
	artistIds: Set<number>;
	genreGenrePairs: Set<string>;
	artistArtistPairs: Set<string>;
	genreArtistEdgeCount: number;
}

// Pre-indexes artist-to-genre relationships once so sliding the 14-day window
// only checks `rangesOverlap` on artists + artist-artist edges, deriving genre
// and genre-genre visibility in O(1) lookups without shipping redundant ranges.
export function buildWindowIndex(
	artists: { artist_id: number; ranges: TimeRange[] }[],
	genreArtistLinks: { genre_id: number; artist_id: number }[],
	genreGenreEdges: { source_genre_id: number; target_genre_id: number }[],
	artistArtistEdges: {
		source_artist_id: number;
		target_artist_id: number;
		ranges: TimeRange[];
	}[],
): WindowIndex {
	const genreArtists = new Map<number, number[]>();
	const genreArtistSets = new Map<number, Set<number>>();
	for (const link of genreArtistLinks) {
		let list = genreArtists.get(link.genre_id);
		if (!list) {
			list = [];
			genreArtists.set(link.genre_id, list);
		}
		list.push(link.artist_id);

		let set = genreArtistSets.get(link.genre_id);
		if (!set) {
			set = new Set<number>();
			genreArtistSets.set(link.genre_id, set);
		}
		set.add(link.artist_id);
	}

	const genreGenreSharedArtists = genreGenreEdges.map((e) => {
		const srcArtists = genreArtists.get(e.source_genre_id) ?? [];
		const tgtSet = genreArtistSets.get(e.target_genre_id);
		const sharedArtistIds = tgtSet
			? srcArtists.filter((aid) => tgtSet.has(aid))
			: [];
		return {
			pairKey: edgePairKey(e.source_genre_id, e.target_genre_id),
			sourceGenreId: e.source_genre_id,
			targetGenreId: e.target_genre_id,
			sharedArtistIds,
		};
	});

	return {
		artists,
		genreArtists,
		genreGenreSharedArtists,
		artistArtistEdges: artistArtistEdges.map((e) => ({
			pairKey: edgePairKey(e.source_artist_id, e.target_artist_id),
			sourceArtistId: e.source_artist_id,
			targetArtistId: e.target_artist_id,
			ranges: e.ranges,
		})),
	};
}

export function evaluateWindow(
	index: WindowIndex,
	windowStart: number,
	windowEnd: number,
): WindowVisibility {
	const artistIds = new Set<number>();
	for (const a of index.artists) {
		if (rangesOverlap(a.ranges, windowStart, windowEnd)) {
			artistIds.add(a.artist_id);
		}
	}

	const genreIds = new Set<number>();
	let genreArtistEdgeCount = 0;
	for (const [genreId, linkedArtists] of index.genreArtists) {
		let activeForGenre = 0;
		for (const aid of linkedArtists) {
			if (artistIds.has(aid)) activeForGenre++;
		}
		if (activeForGenre > 0) {
			genreIds.add(genreId);
			genreArtistEdgeCount += activeForGenre;
		}
	}

	const genreGenrePairs = new Set<string>();
	for (const e of index.genreGenreSharedArtists) {
		if (
			genreIds.has(e.sourceGenreId) &&
			genreIds.has(e.targetGenreId) &&
			e.sharedArtistIds.some((aid) => artistIds.has(aid))
		) {
			genreGenrePairs.add(e.pairKey);
		}
	}

	const artistArtistPairs = new Set<string>();
	for (const e of index.artistArtistEdges) {
		if (
			artistIds.has(e.sourceArtistId) &&
			artistIds.has(e.targetArtistId) &&
			rangesOverlap(e.ranges, windowStart, windowEnd)
		) {
			artistArtistPairs.add(e.pairKey);
		}
	}

	return {
		genreIds,
		artistIds,
		genreGenrePairs,
		artistArtistPairs,
		genreArtistEdgeCount,
	};
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

// Kind glyph drawn centered on a node by the galaxy graph's overlay canvas.
// Sigma 4's native labels are monochrome SDF text, so color emoji can only
// come from our own canvas layer.
export function kindGlyph(kind: "genre" | "artist"): string {
	return kind === "artist" ? "🎨" : "🎵";
}

// Glyph font size scales with the node but never drops below a legible floor.
export function glyphSize(renderedNodeSize: number): number {
	return Math.max(8, renderedNodeSize * 1.2);
}

// Whether the overlay draws a node's glyph + label. Mirrors Sigma 3's
// `labelRenderedSizeThreshold` (small nodes stay unlabeled until zoomed in),
// except the hovered node always gets drawn so its pill is never missing.
export function shouldDrawLabel(
	renderedNodeSize: number,
	threshold: number,
	isHovered: boolean,
): boolean {
	return isHovered || renderedNodeSize >= threshold;
}

// Thins a set of label candidates so no grid cell holds more than `perCell`
// labels, keeping the largest nodes in each cell. Stands in for Sigma 3's
// label grid (`labelGridCellSize`/`labelDensity`), which native Sigma 4 labels
// have but our overlay canvas does not get for free. Input order is not
// preserved: the result is sorted largest-first.
export function cullLabelsByGrid<
	T extends { x: number; y: number; size: number },
>(items: T[], cellSize: number, perCell: number): T[] {
	const counts = new Map<string, number>();
	const kept: T[] = [];
	for (const item of [...items].sort((a, b) => b.size - a.size)) {
		const key = `${Math.floor(item.x / cellSize)}:${Math.floor(item.y / cellSize)}`;
		const used = counts.get(key) ?? 0;
		if (used >= perCell) continue;
		counts.set(key, used + 1);
		kept.push(item);
	}
	return kept;
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
// one). Raising t to the 4th power pushes weak/mid-weight edges close to
// invisible, which matters because with thousands of overlapping edges, alpha
// compositing stacks even "low" per-edge opacity into a bright haze.
//
// The range (0.15–0.5) is tuned for Sigma 4's native edge `opacity` style,
// which composites more faintly than the alpha channel Sigma 3's curved edge
// program read from an rgba() color; the 0.01–0.12 range that looked right
// there renders as no visible edges at all.
export function edgeOpacity(count: number, maxCount: number): number {
	const t = maxCount > 0 ? count / maxCount : 0;
	return 0.15 + 0.35 * t * t * t * t;
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

// Deterministic Mulberry32 PRNG returning values in [0, 1). Keeps community
// detection and initial circle-packing layouts reproducible across reloads.
export function createSeededRng(seed: number): () => number {
	let state = seed >>> 0;
	return () => {
		state = (state + 0x6d2b79f5) >>> 0;
		let t = Math.imul(state ^ (state >>> 15), 1 | state);
		t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
		return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
	};
}

// Log-compresses positive edge weights onto [0.15, 1.0] against the strongest
// edge of the same kind so genre-genre, genre-artist, and artist-artist links
// participate on a commensurate scale during clustering and force layout.
export function logNormalizedWeight(weight: number, maxWeight: number): number {
	if (weight <= 0 || maxWeight <= 0) return 0;
	const ratio = Math.min(1, Math.log1p(weight) / Math.log1p(maxWeight));
	return 0.15 + 0.85 * ratio;
}

// Retains the union of each endpoint's top-K heaviest incident edges (by index
// into `edges`). Every node with degree > 0 keeps at least one edge while
// filtering dense cross-cluster hairball noise out of the resting backbone.
export function selectBackboneEdges<T>(
	edges: readonly T[],
	getSource: (edge: T) => string,
	getTarget: (edge: T) => string,
	getWeight: (edge: T) => number,
	topKPerEndpoint: number,
): Set<number> {
	if (topKPerEndpoint <= 0 || edges.length === 0) return new Set();

	const incidentByNode = new Map<string, { index: number; weight: number }[]>();
	const pushIncident = (node: string, index: number, weight: number): void => {
		let list = incidentByNode.get(node);
		if (!list) {
			list = [];
			incidentByNode.set(node, list);
		}
		list.push({ index, weight });
	};

	for (let i = 0; i < edges.length; i++) {
		const edge = edges[i];
		const w = getWeight(edge);
		pushIncident(getSource(edge), i, w);
		pushIncident(getTarget(edge), i, w);
	}

	const kept = new Set<number>();
	for (const list of incidentByNode.values()) {
		list.sort((a, b) => b.weight - a.weight || a.index - b.index);
		const limit = Math.min(topKPerEndpoint, list.length);
		for (let i = 0; i < limit; i++) {
			kept.add(list[i].index);
		}
	}
	return kept;
}

/**
 * Assigns palette colors to communities ordered from largest to smallest
 * member count (ties broken by community ID ascending), reserving vivid
 * palette hues for substantive clusters (`size >= minMembersForPalette`) and
 * assigning tiny islands a muted neutral color.
 */
export function assignCommunityColorsBySize(
	communitySizes: ReadonlyMap<number, number>,
	palette: readonly string[],
	minMembersForPalette = 3,
	smallCommunityColor = "#5a5a78",
): Map<number, string> {
	const entries = [...communitySizes.entries()].sort(
		(a, b) => b[1] - a[1] || a[0] - b[0],
	);
	const result = new Map<number, string>();
	let paletteIdx = 0;
	for (const [commId, count] of entries) {
		if (count < minMembersForPalette || palette.length === 0) {
			result.set(commId, smallCommunityColor);
		} else {
			result.set(commId, palette[paletteIdx % palette.length]);
			paletteIdx++;
		}
	}
	return result;
}

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
