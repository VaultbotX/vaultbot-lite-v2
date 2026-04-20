export function communityColor(commId: number, numCommunities: number): string {
	const hue = Math.round(((commId / numCommunities) * 360 + 200) % 360);
	return `hsl(${hue}, 62%, 56%)`;
}

// Log-scale node diameter: 18–58 px
export function nodeSize(count: number, maxCount: number): number {
	return 18 + 40 * (Math.log(count + 1) / Math.log(maxCount + 1));
}

// Linear edge width: 0.5–4 px
export function edgeWidth(count: number, maxShared: number): number {
	return 0.5 + 3.5 * (count / maxShared);
}
