export function communityColor(commId: number, numCommunities: number): string {
	const hue = Math.round(((commId / numCommunities) * 360 + 200) % 360);
	return `hsl(${hue}, 62%, 56%)`;
}

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
