// Log-scale node diameter
export function nodeSize(count: number, maxCount: number): number {
	return 14 + 50 * (Math.log(count + 1) / Math.log(maxCount + 1));
}

// Sqrt-scale edge width
export function edgeWidth(count: number, maxShared: number): number {
	return 0.5 + 5 * Math.sqrt(count / maxShared);
}

// Sqrt-scale edge opacity, used to fade thin edges relative to the strongest
// edge of the same kind (edge kinds live on different weight scales, so each
// kind must be normalized against its own max, not a global one).
export function edgeOpacity(count: number, maxCount: number): number {
	return 0.15 + 0.5 * Math.sqrt(maxCount > 0 ? count / maxCount : 0);
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

export interface SelectedNode {
	kind: "genre" | "artist";
	id: number;
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
