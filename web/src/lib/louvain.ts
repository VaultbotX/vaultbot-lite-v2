export class CommunityPartition {
	private readonly _nodeToCommunity: ReadonlyMap<number, number>;
	private readonly _communityToNodes: ReadonlyMap<number, ReadonlySet<number>>;

	constructor(nodeToCommunity: Map<number, number>) {
		this._nodeToCommunity = nodeToCommunity;
		const c2n = new Map<number, Set<number>>();
		for (const [nodeId, commId] of nodeToCommunity) {
			if (!c2n.has(commId)) c2n.set(commId, new Set());
			c2n.get(commId)!.add(nodeId);
		}
		this._communityToNodes = c2n;
	}

	communityOf(nodeId: number): number | undefined {
		return this._nodeToCommunity.get(nodeId);
	}

	membersOf(communityId: number): ReadonlySet<number> {
		return this._communityToNodes.get(communityId) ?? new Set();
	}

	get communityIds(): readonly number[] {
		return [...this._communityToNodes.keys()];
	}

	get communityCount(): number { return this._communityToNodes.size; }
	get nodeCount(): number { return this._nodeToCommunity.size; }
	isEmpty(): boolean { return this._nodeToCommunity.size === 0; }
}

// Louvain community detection (phase 1 greedy modularity maximization)
// Returns a CommunityPartition with nodeId → community index (0-based consecutive integers)
export function detectCommunities(
	nodeIds: number[],
	edges: { source: number; target: number; weight: number }[],
): CommunityPartition {
	const n = nodeIds.length;
	if (n === 0) return new CommunityPartition(new Map());

	const idx = new Map<number, number>(nodeIds.map((id, i) => [id, i]));

	const adj: Map<number, number>[] = Array.from({ length: n }, () => new Map());
	let m = 0;

	for (const { source, target, weight } of edges) {
		const s = idx.get(source);
		const t = idx.get(target);
		if (s === undefined || t === undefined || s === t) continue;
		adj[s].set(t, (adj[s].get(t) ?? 0) + weight);
		adj[t].set(s, (adj[t].get(s) ?? 0) + weight);
		m += weight;
	}

	if (m === 0) return new CommunityPartition(new Map(nodeIds.map((id, i) => [id, i])));

	// k[i] = sum of edge weights incident to node i
	const k: number[] = nodeIds.map((_, i) => {
		let sum = 0;
		for (const w of adj[i].values()) sum += w;
		return sum;
	});

	// Each node starts in its own community
	const comm: number[] = nodeIds.map((_, i) => i);
	// sigma_tot[c] = sum of k[i] for all i in community c
	const sigma_tot: number[] = [...k];

	let improved = true;
	let pass = 0;
	while (improved && pass < 20) {
		improved = false;
		pass++;
		for (let i = 0; i < n; i++) {
			const ci = comm[i];

			// k_i_ci: edges from i to other nodes already in ci
			let k_i_ci = 0;
			for (const [j, w] of adj[i]) {
				if (comm[j] === ci) k_i_ci += w;
			}

			// Collect neighbour communities and edge weight from i to each
			const neighComm = new Map<number, number>();
			for (const [j, w] of adj[i]) {
				const cj = comm[j];
				if (cj !== ci) neighComm.set(cj, (neighComm.get(cj) ?? 0) + w);
			}

			// Modularity gain of removing i from ci (Blondel et al. 2008)
			const remove_gain =
				k_i_ci / m - ((sigma_tot[ci] - k[i]) * k[i]) / (2 * m * m);

			let bestDQ = 0;
			let bestC = ci;
			for (const [cj, k_i_cj] of neighComm) {
				const add_gain = k_i_cj / m - (sigma_tot[cj] * k[i]) / (2 * m * m);
				const dq = add_gain - remove_gain;
				if (dq > bestDQ) {
					bestDQ = dq;
					bestC = cj;
				}
			}

			if (bestC !== ci) {
				sigma_tot[ci] -= k[i];
				sigma_tot[bestC] += k[i];
				comm[i] = bestC;
				improved = true;
			}
		}
	}

	// Remap community IDs to consecutive integers ordered by first appearance
	const seen = new Map<number, number>();
	let nextId = 0;
	return new CommunityPartition(
		new Map(
			nodeIds.map((id, i) => {
				const c = comm[i];
				if (!seen.has(c)) seen.set(c, nextId++);
				return [id, seen.get(c) ?? 0];
			}),
		),
	);
}
