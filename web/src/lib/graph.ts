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

// Eliminates remaining label-on-label collisions that straddle adjacent spatial
// grid boundaries. Expects `items` ordered highest-priority (largest) first and
// returns non-overlapping winners in that same priority order.
export function cullOverlappingLabels<
	T extends { x: number; y: number; size: number; label: string },
>(items: readonly T[], fontSize = 12, padX = 6, padY = 3): T[] {
	const kept: T[] = [];
	const boxes: { x0: number; y0: number; x1: number; y1: number }[] = [];
	const charWidth = fontSize * 0.58;
	const halfH = fontSize * 0.5 + padY;

	for (const item of items) {
		const x0 = item.x - item.size - 2;
		const x1 = item.x + item.size + 4 + item.label.length * charWidth + padX;
		const y0 = item.y - halfH;
		const y1 = item.y + halfH;

		let collides = false;
		for (const b of boxes) {
			if (x0 < b.x1 && x1 > b.x0 && y0 < b.y1 && y1 > b.y0) {
				collides = true;
				break;
			}
		}
		if (!collides) {
			kept.push(item);
			boxes.push({ x0, y0, x1, y1 });
		}
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

// Spatial-grid circle collision pass. Separates overlapping nodes in place
// while weighting displacement inversely by radius so large genre hubs stay
// anchored near their constellation's core and smaller orbiting nodes give way.
export function separateCircles(
	nodes: { x: number; y: number }[],
	radii: readonly number[],
	iterations = 14,
): void {
	if (nodes.length < 2 || iterations <= 0) return;

	let maxRadius = 1;
	for (const r of radii) {
		if (r > maxRadius) maxRadius = r;
	}
	const cellSize = Math.max(2 * maxRadius, 1);
	const GOLDEN_ANGLE = Math.PI * (3 - Math.sqrt(5));

	for (let iter = 0; iter < iterations; iter++) {
		const grid = new Map<string, number[]>();
		for (let i = 0; i < nodes.length; i++) {
			const cx = Math.floor(nodes[i].x / cellSize);
			const cy = Math.floor(nodes[i].y / cellSize);
			const key = `${cx}:${cy}`;
			let bucket = grid.get(key);
			if (!bucket) {
				bucket = [];
				grid.set(key, bucket);
			}
			bucket.push(i);
		}

		let moved = false;
		for (let i = 0; i < nodes.length; i++) {
			const ni = nodes[i];
			const ri = radii[i];
			const cx = Math.floor(ni.x / cellSize);
			const cy = Math.floor(ni.y / cellSize);

			for (let dx = -1; dx <= 1; dx++) {
				for (let dy = -1; dy <= 1; dy++) {
					const bucket = grid.get(`${cx + dx}:${cy + dy}`);
					if (!bucket) continue;
					for (const j of bucket) {
						if (j <= i) continue;
						const nj = nodes[j];
						const minDist = ri + radii[j];
						if (minDist <= 0) continue;
						let vx = nj.x - ni.x;
						let vy = nj.y - ni.y;
						const distSq = vx * vx + vy * vy;
						if (distSq >= minDist * minDist) continue;

						moved = true;
						let dist = Math.sqrt(distSq);
						if (dist < 1e-6) {
							const angle = ((i * 7 + j * 13) % 360) * GOLDEN_ANGLE;
							vx = Math.cos(angle);
							vy = Math.sin(angle);
							dist = 1;
						} else {
							vx /= dist;
							vy /= dist;
						}

						const overlap = (minDist - dist) * 0.7;
						const wi = radii[j] / minDist;
						const wj = ri / minDist;
						ni.x -= vx * overlap * wi;
						ni.y -= vy * overlap * wi;
						nj.x += vx * overlap * wj;
						nj.y += vy * overlap * wj;
					}
				}
			}
		}
		if (!moved) break;
	}
}

export interface ConstellationLayoutOptions {
	aspectX?: number;
	channelGap?: number;
}

/**
 * Positions macro-constellation centers across a widescreen galactic plane:
 * 1. Seeds constellations along a greedy-affinity spiral so strongly bridged
 *    communities start near one another.
 * 2. Relaxes free topological springs + soft repulsion without collision walls
 *    so related musical clusters find their natural neighborhoods.
 * 3. Rotates the principal axis horizontally (PCA) to fit widescreen viewports.
 * 4. Enforces a hard `R_i + R_j + channelGap` separation so every pair of
 *    constellations is divided by visible dark space.
 */
export function layoutConstellationCenters(
	radii: readonly number[],
	affinities: readonly (readonly number[])[],
	options: ConstellationLayoutOptions = {},
): { x: number; y: number }[] {
	const count = radii.length;
	if (count === 0) return [];
	if (count === 1) return [{ x: 0, y: 0 }];

	const aspectX = options.aspectX ?? 1.85;
	const channelGap = options.channelGap ?? 54;

	let maxAffinity = 1e-6;
	for (let i = 0; i < count; i++) {
		for (let j = i + 1; j < count; j++) {
			const w = affinities[i]?.[j] ?? 0;
			if (w > maxAffinity) maxAffinity = w;
		}
	}

	const placedOrder = [0];
	const used = new Set([0]);
	while (placedOrder.length < count) {
		let bestNext = -1;
		let bestScore = -1;
		const last = placedOrder[placedOrder.length - 1];
		for (let j = 0; j < count; j++) {
			if (used.has(j)) continue;
			const score =
				(affinities[last]?.[j] ?? 0) * 2 + (affinities[0]?.[j] ?? 0) * 0.5;
			if (score > bestScore) {
				bestScore = score;
				bestNext = j;
			}
		}
		placedOrder.push(bestNext);
		used.add(bestNext);
	}

	const GOLDEN_ANGLE = Math.PI * (3 - Math.sqrt(5));
	const coords = Array.from({ length: count }, () => ({ x: 0, y: 0 }));
	for (let rank = 0; rank < count; rank++) {
		const idx = placedOrder[rank];
		const r = rank === 0 ? 0 : 90 * Math.sqrt(rank);
		const theta = rank * GOLDEN_ANGLE;
		coords[idx].x = r * Math.cos(theta);
		coords[idx].y = r * Math.sin(theta);
	}

	// Free isotropic relaxation so related clusters pass unhindered into neighbors.
	for (let iter = 0; iter < 220; iter++) {
		const fx = new Float64Array(count);
		const fy = new Float64Array(count);
		for (let i = 0; i < count; i++) {
			fx[i] -= coords[i].x * 0.025;
			fy[i] -= coords[i].y * 0.025;
			for (let j = i + 1; j < count; j++) {
				const dx = coords[j].x - coords[i].x;
				const dy = coords[j].y - coords[i].y;
				const dist = Math.hypot(dx, dy) || 1;
				const ux = dx / dist;
				const uy = dy / dist;
				const rep = (radii[i] * radii[j] * 0.9) / Math.max(dist, 25);
				fx[i] -= ux * rep;
				fy[i] -= uy * rep;
				fx[j] += ux * rep;
				fy[j] += uy * rep;

				const aff = affinities[i]?.[j] ?? 0;
				if (aff > 0) {
					const normAff = Math.sqrt(aff / maxAffinity);
					const att = dist * 0.085 * normAff;
					fx[i] += ux * att;
					fy[i] += uy * att;
					fx[j] -= ux * att;
					fy[j] -= uy * att;
				}
			}
		}
		for (let i = 0; i < count; i++) {
			coords[i].x += fx[i] * 0.35;
			coords[i].y += fy[i] * 0.35;
		}
	}

	// Align principal axis horizontally along X for widescreen viewports.
	let meanX = 0;
	let meanY = 0;
	let totalWeight = 0;
	for (let i = 0; i < count; i++) {
		const w = Math.max(radii[i] * radii[i], 1);
		meanX += coords[i].x * w;
		meanY += coords[i].y * w;
		totalWeight += w;
	}
	meanX /= totalWeight;
	meanY /= totalWeight;

	let cxx = 0;
	let cyy = 0;
	let cxy = 0;
	for (let i = 0; i < count; i++) {
		coords[i].x -= meanX;
		coords[i].y -= meanY;
		const w = Math.max(radii[i] * radii[i], 1);
		cxx += w * coords[i].x * coords[i].x;
		cyy += w * coords[i].y * coords[i].y;
		cxy += w * coords[i].x * coords[i].y;
	}
	const pcaAngle = -0.5 * Math.atan2(2 * cxy, cxx - cyy);
	const cosA = Math.cos(pcaAngle);
	const sinA = Math.sin(pcaAngle);
	for (const c of coords) {
		const rx = c.x * cosA - c.y * sinA;
		const ry = c.x * sinA + c.y * cosA;
		c.x = rx;
		c.y = ry;
	}
	if (
		coords[0].x < -1e-6 ||
		(Math.abs(coords[0].x) <= 1e-6 && coords[1].x < 0)
	) {
		for (const c of coords) {
			c.x = -c.x;
			c.y = -c.y;
		}
	}
	for (const c of coords) {
		c.x *= 1.5;
		c.y *= 1.35 / aspectX;
	}

	const pushApart = (strength: number, horizontalBias = 1.0): void => {
		for (let i = 0; i < count; i++) {
			for (let j = i + 1; j < count; j++) {
				let dx = coords[j].x - coords[i].x;
				let dy = coords[j].y - coords[i].y;
				let dist = Math.hypot(dx, dy);
				const minSep = radii[i] + radii[j] + channelGap;
				if (dist >= minSep) continue;
				if (dist < 1e-4) {
					dx = Math.cos(i * 5 + j * 11);
					dy = Math.sin(i * 5 + j * 11);
					dist = 1;
				}
				const push = (minSep - dist) * strength;
				let ux = (dx / dist) * horizontalBias;
				let uy = dy / dist / horizontalBias;
				const uLen = Math.hypot(ux, uy) || 1;
				ux /= uLen;
				uy /= uLen;
				const denom = Math.max(radii[i] + radii[j], 1);
				const wi = radii[j] / denom;
				const wj = radii[i] / denom;
				coords[i].x -= ux * push * wi;
				coords[i].y -= uy * push * wi;
				coords[j].x += ux * push * wj;
				coords[j].y += uy * push * wj;
			}
		}
	};

	// Damped affinity + widescreen band pull + cosmic-channel separation.
	for (let iter = 0; iter < 220; iter++) {
		const fx = new Float64Array(count);
		const fy = new Float64Array(count);
		const alpha = 1 - iter / 250;

		for (let i = 0; i < count; i++) {
			fx[i] -= coords[i].x * 0.003;
			fy[i] -= coords[i].y * 0.055;

			for (let j = i + 1; j < count; j++) {
				const aff = affinities[i]?.[j] ?? 0;
				if (aff <= 0) continue;
				const dx = coords[j].x - coords[i].x;
				const dy = coords[j].y - coords[i].y;
				const dist = Math.hypot(dx, dy) || 1;
				const minSep = radii[i] + radii[j] + channelGap;
				const normAff = Math.sqrt(aff / maxAffinity);
				const restLen = minSep + (1 - normAff) * 65;
				const spring = (dist - restLen) * (0.036 * normAff);
				fx[i] += (dx / dist) * spring;
				fy[i] += (dy / dist) * spring;
				fx[j] -= (dx / dist) * spring;
				fy[j] -= (dy / dist) * spring;
			}
		}

		for (let i = 0; i < count; i++) {
			coords[i].x += fx[i] * alpha;
			coords[i].y += fy[i] * alpha;
		}
		for (let sub = 0; sub < 4; sub++) pushApart(0.65, 1.25);
	}

	// Final hard separation pass guaranteeing every cosmic channel >= channelGap.
	for (let sub = 0; sub < 14; sub++) pushApart(1.0, 1.0);

	let minX = Number.POSITIVE_INFINITY;
	let maxX = Number.NEGATIVE_INFINITY;
	let minY = Number.POSITIVE_INFINITY;
	let maxY = Number.NEGATIVE_INFINITY;
	for (let i = 0; i < count; i++) {
		if (coords[i].x - radii[i] < minX) minX = coords[i].x - radii[i];
		if (coords[i].x + radii[i] > maxX) maxX = coords[i].x + radii[i];
		if (coords[i].y - radii[i] < minY) minY = coords[i].y - radii[i];
		if (coords[i].y + radii[i] > maxY) maxY = coords[i].y + radii[i];
	}
	const midX = (minX + maxX) / 2;
	const midY = (minY + maxY) / 2;
	for (const c of coords) {
		c.x -= midX;
		c.y -= midY;
	}

	return coords;
}

export interface ExternalPull {
	nodeId: string;
	dirX: number;
	dirY: number;
	weight: number;
}

// Analytical 2D Procrustes angle maximizing alignment between cross-community
// bridge nodes inside a constellation and the directions of the neighboring
// constellations they connect to.
export function optimalConstellationRotation(
	localNodes: readonly { id: string; x: number; y: number }[],
	externalPulls: readonly ExternalPull[],
): number {
	if (localNodes.length === 0 || externalPulls.length === 0) return 0;
	const byId = new Map<string, { x: number; y: number }>();
	for (const n of localNodes) byId.set(n.id, n);

	let a = 0;
	let b = 0;
	for (const pull of externalPulls) {
		if (pull.weight <= 0) continue;
		const pos = byId.get(pull.nodeId);
		if (!pos) continue;
		const len = Math.hypot(pull.dirX, pull.dirY);
		if (len < 1e-6) continue;
		const ux = pull.dirX / len;
		const uy = pull.dirY / len;
		a += pull.weight * (pos.x * ux + pos.y * uy);
		b += pull.weight * (pos.x * uy - pos.y * ux);
	}
	return a * a + b * b > 1e-9 ? Math.atan2(b, a) : 0;
}

// Places disconnected micro-clusters and isolated singleton nodes along an
// elliptical outer belt hugging the main galaxy's bounding box so repulsion
// never flings them into deep space or distorts widescreen framing.
export function placeSatelliteComponents(
	corePositions: readonly { x: number; y: number }[],
	satellites: readonly (readonly string[])[],
): Map<string, { x: number; y: number }> {
	const result = new Map<string, { x: number; y: number }>();
	if (satellites.length === 0) return result;

	let minX = Number.POSITIVE_INFINITY;
	let maxX = Number.NEGATIVE_INFINITY;
	let minY = Number.POSITIVE_INFINITY;
	let maxY = Number.NEGATIVE_INFINITY;
	for (const p of corePositions) {
		if (p.x < minX) minX = p.x;
		if (p.x > maxX) maxX = p.x;
		if (p.y < minY) minY = p.y;
		if (p.y > maxY) maxY = p.y;
	}

	const cx = corePositions.length > 0 ? (minX + maxX) / 2 : 0;
	const cy = corePositions.length > 0 ? (minY + maxY) / 2 : 0;
	const rx = corePositions.length > 0 ? Math.max((maxX - minX) / 2, 50) : 100;
	const ry = corePositions.length > 0 ? Math.max((maxY - minY) / 2, 50) : 100;
	const localClusterRadius = Math.min(rx, ry) * 0.016;

	const sorted = [...satellites].sort((a, b) => b.length - a.length);
	for (let i = 0; i < sorted.length; i++) {
		const comp = sorted[i];
		const angle = (2 * Math.PI * i) / sorted.length;
		const beltScale = 1.07 + ((i % 3) / 2) * 0.05;
		const ax = cx + rx * beltScale * Math.cos(angle);
		const ay = cy + ry * beltScale * Math.sin(angle);

		if (comp.length === 1) {
			result.set(comp[0], { x: ax, y: ay });
		} else {
			for (let m = 0; m < comp.length; m++) {
				const ma = (2 * Math.PI * m) / comp.length;
				result.set(comp[m], {
					x: ax + localClusterRadius * Math.cos(ma),
					y: ay + localClusterRadius * Math.sin(ma),
				});
			}
		}
	}

	return result;
}

export interface NeighborhoodCameraInput {
	points: readonly { x: number; y: number }[];
	viewport: { width: number; height: number };
	graphDimensions: { width: number; height: number };
	stagePadding: number;
	padding?: { top: number; right: number; bottom: number; left: number };
	singlePointRatio?: number;
	minRatio?: number;
	maxRatio?: number;
}

// Mirrors Sigma 4's internal `getCorrectionRatio` so camera bounding-box fits
// match Sigma's `matrixFromCamera` projection across any stage/graph aspect.
export function sigmaCorrectionRatio(
	viewport: { width: number; height: number },
	graphDimensions: { width: number; height: number },
): number {
	const viewportRatio = viewport.height / viewport.width;
	const graphRatio = graphDimensions.height / graphDimensions.width;
	if (
		(viewportRatio < 1 && graphRatio > 1) ||
		(viewportRatio > 1 && graphRatio < 1)
	) {
		return 1;
	}
	return Math.min(
		Math.max(graphRatio, 1 / graphRatio),
		Math.max(1 / viewportRatio, viewportRatio),
	);
}

/**
 * Computes the Sigma camera `{ x, y, ratio }` state (in framedGraph coordinates)
 * that centers and fits a selected node together with all of its connected
 * neighbors inside `viewport`, reserving extra right-hand padding so rightward
 * node labels and focus pills aren't clipped by the detail drawer.
 */
export function computeNeighborhoodCameraState(
	input: NeighborhoodCameraInput,
): { x: number; y: number; ratio: number } | null {
	const { points } = input;
	if (points.length === 0) return null;

	const pad = input.padding ?? { top: 68, right: 120, bottom: 68, left: 68 };
	const singlePointRatio = input.singlePointRatio ?? 0.15;
	const minRatio = input.minRatio ?? 0.14;
	const maxRatio = input.maxRatio ?? 1.6;

	const width = Math.max(1, input.viewport.width);
	const height = Math.max(1, input.viewport.height);
	const gw = Math.max(1e-6, input.graphDimensions.width);
	const gh = Math.max(1e-6, input.graphDimensions.height);

	const smallestDimension = Math.max(
		1,
		Math.min(width, height) - 2 * input.stagePadding,
	);
	const correction = sigmaCorrectionRatio(
		{ width, height },
		{ width: gw, height: gh },
	);
	const baseScale = smallestDimension * correction;

	let minX = Number.POSITIVE_INFINITY;
	let maxX = Number.NEGATIVE_INFINITY;
	let minY = Number.POSITIVE_INFINITY;
	let maxY = Number.NEGATIVE_INFINITY;
	for (const p of points) {
		if (p.x < minX) minX = p.x;
		if (p.x > maxX) maxX = p.x;
		if (p.y < minY) minY = p.y;
		if (p.y > maxY) maxY = p.y;
	}

	const midX = (minX + maxX) / 2;
	const midY = (minY + maxY) / 2;
	const spanX = maxX - minX;
	const spanY = maxY - minY;

	let ratio: number;
	if (points.length === 1 || (spanX < 1e-6 && spanY < 1e-6)) {
		ratio = Math.min(maxRatio, Math.max(minRatio, singlePointRatio));
	} else {
		const availWidth = Math.max(80, width - pad.left - pad.right);
		const availHeight = Math.max(80, height - pad.top - pad.bottom);
		const ratioX = (baseScale * spanX) / availWidth;
		const ratioY = (baseScale * spanY) / availHeight;
		ratio = Math.min(maxRatio, Math.max(minRatio, Math.max(ratioX, ratioY)));
	}

	return {
		x: midX + ((pad.right - pad.left) / 2) * (ratio / baseScale),
		y: midY + ((pad.top - pad.bottom) / 2) * (ratio / baseScale),
		ratio,
	};
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

// 20 vivid hues so all macro-constellations in a multi-year archive receive
// distinct, high-contrast colors on dark `--bg`. Hex values (not HSL) —
// sigma's WebGL renderer parses hex/rgb only.
export const COMMUNITY_PALETTE: readonly string[] = [
	"#DA654E", // coral
	"#4EDAAB", // mint
	"#DA4EC2", // magenta
	"#7DDA4E", // lime
	"#654EDA", // violet
	"#DAAB4E", // gold
	"#4EC2DA", // sky
	"#DA4E7D", // rose
	"#4EDA65", // emerald
	"#AB4EDA", // purple
	"#C2DA4E", // chartreuse
	"#4E7DDA", // cornflower
	"#DA8D4E", // amber
	"#4E96DA", // azure
	"#B84EDA", // orchid
	"#4EDAC8", // turquoise
	"#DA4E58", // crimson
	"#9FDA4E", // spring green
	"#7D65DA", // indigo
	"#DAC84E", // sunflower
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
