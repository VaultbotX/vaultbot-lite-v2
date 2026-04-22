// Log-scale node diameter
export function nodeSize(count: number, maxCount: number): number {
	return 14 + 50 * (Math.log(count + 1) / Math.log(maxCount + 1));
}

// Sqrt-scale edge width
export function edgeWidth(count: number, maxShared: number): number {
	return 0.5 + 5 * Math.sqrt(count / maxShared);
}

// Ideal edge length for fcose layout: shorter for densely-shared edges
export function idealEdgeLength(shared: number): number {
	return Math.max(50, 150 / Math.sqrt(shared || 1));
}

// Edge elasticity for fcose layout: stiffer for densely-shared edges
export function edgeElasticity(shared: number): number {
	return Math.min(0.9, 0.05 + (shared || 1) / 12);
}

// 12 hues stepped by 150° so consecutive community IDs look maximally different
export const COMMUNITY_PALETTE: readonly string[] = [
	"hsl(10, 65%, 58%)",
	"hsl(160, 65%, 58%)",
	"hsl(310, 65%, 58%)",
	"hsl(100, 65%, 58%)",
	"hsl(250, 65%, 58%)",
	"hsl(40, 65%, 58%)",
	"hsl(190, 65%, 58%)",
	"hsl(340, 65%, 58%)",
	"hsl(130, 65%, 58%)",
	"hsl(280, 65%, 58%)",
	"hsl(70, 65%, 58%)",
	"hsl(220, 65%, 58%)",
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
