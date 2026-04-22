import { describe, expect, it } from "vitest";
import { CommunityPartition, detectCommunities } from "./louvain";

describe("CommunityPartition", () => {
	it("communityOf returns the community for a known node", () => {
		const p = new CommunityPartition(new Map([[1, 0], [2, 1]]));
		expect(p.communityOf(1)).toBe(0);
		expect(p.communityOf(2)).toBe(1);
	});

	it("communityOf returns undefined for an unknown node", () => {
		const p = new CommunityPartition(new Map([[1, 0]]));
		expect(p.communityOf(99)).toBeUndefined();
	});

	it("membersOf returns all nodes in a community", () => {
		const p = new CommunityPartition(new Map([[1, 0], [2, 0], [3, 1]]));
		expect(p.membersOf(0)).toEqual(new Set([1, 2]));
		expect(p.membersOf(1)).toEqual(new Set([3]));
	});

	it("membersOf returns empty set for unknown community", () => {
		const p = new CommunityPartition(new Map([[1, 0]]));
		expect(p.membersOf(99).size).toBe(0);
	});

	it("communityIds lists all unique community IDs", () => {
		const p = new CommunityPartition(new Map([[1, 0], [2, 1], [3, 0]]));
		expect(new Set(p.communityIds)).toEqual(new Set([0, 1]));
	});

	it("communityCount equals number of unique communities", () => {
		const p = new CommunityPartition(new Map([[1, 0], [2, 1], [3, 0]]));
		expect(p.communityCount).toBe(2);
	});

	it("nodeCount equals total number of nodes", () => {
		const p = new CommunityPartition(new Map([[1, 0], [2, 1], [3, 0]]));
		expect(p.nodeCount).toBe(3);
	});

	it("isEmpty returns true for empty partition", () => {
		expect(new CommunityPartition(new Map()).isEmpty()).toBe(true);
	});

	it("isEmpty returns false for non-empty partition", () => {
		expect(new CommunityPartition(new Map([[1, 0]])).isEmpty()).toBe(false);
	});
});

describe("detectCommunities", () => {
	it("returns empty partition for empty input", () => {
		expect(detectCommunities([], []).isEmpty()).toBe(true);
	});

	it("assigns each isolated node its own community when there are no edges", () => {
		const result = detectCommunities([1, 2, 3], []);
		expect(result.nodeCount).toBe(3);
		expect(result.communityCount).toBe(3);
	});

	it("places two directly connected nodes in the same community", () => {
		const result = detectCommunities([1, 2], [{ source: 1, target: 2, weight: 1 }]);
		expect(result.communityOf(1)).toBe(result.communityOf(2));
	});

	it("merges a fully-connected triangle into one community", () => {
		const edges = [
			{ source: 1, target: 2, weight: 1 },
			{ source: 2, target: 3, weight: 1 },
			{ source: 1, target: 3, weight: 1 },
		];
		const result = detectCommunities([1, 2, 3], edges);
		expect(result.communityOf(1)).toBe(result.communityOf(2));
		expect(result.communityOf(2)).toBe(result.communityOf(3));
	});

	it("separates two dense triangles connected by a weak bridge", () => {
		const edges = [
			{ source: 1, target: 2, weight: 10 },
			{ source: 2, target: 3, weight: 10 },
			{ source: 1, target: 3, weight: 10 },
			{ source: 4, target: 5, weight: 10 },
			{ source: 5, target: 6, weight: 10 },
			{ source: 4, target: 6, weight: 10 },
			{ source: 3, target: 4, weight: 1 },
		];
		const result = detectCommunities([1, 2, 3, 4, 5, 6], edges);
		expect(result.communityOf(1)).toBe(result.communityOf(2));
		expect(result.communityOf(2)).toBe(result.communityOf(3));
		expect(result.communityOf(4)).toBe(result.communityOf(5));
		expect(result.communityOf(5)).toBe(result.communityOf(6));
		expect(result.communityOf(1)).not.toBe(result.communityOf(4));
	});

	it("returns 0-based consecutive community IDs", () => {
		const edges = [
			{ source: 1, target: 2, weight: 10 },
			{ source: 2, target: 3, weight: 10 },
			{ source: 1, target: 3, weight: 10 },
			{ source: 4, target: 5, weight: 10 },
			{ source: 5, target: 6, weight: 10 },
			{ source: 4, target: 6, weight: 10 },
		];
		const result = detectCommunities([1, 2, 3, 4, 5, 6], edges);
		const ids = [...new Set(result.communityIds)].sort((a, b) => a - b);
		expect(ids[0]).toBe(0);
		for (let i = 1; i < ids.length; i++) {
			expect(ids[i]).toBe(ids[i - 1] + 1);
		}
	});

	it("ignores self-loop edges", () => {
		const result = detectCommunities([1, 2], [{ source: 1, target: 1, weight: 5 }]);
		expect(result.nodeCount).toBe(2);
	});

	it("handles nodes not referenced by any edge", () => {
		const edges = [{ source: 1, target: 2, weight: 5 }];
		const result = detectCommunities([1, 2, 3], edges);
		expect(result.nodeCount).toBe(3);
		expect(result.communityOf(1)).toBe(result.communityOf(2));
		expect(result.communityOf(3)).not.toBe(result.communityOf(1));
	});

	it("handles edges with zero weight without crashing", () => {
		const result = detectCommunities([1, 2], [{ source: 1, target: 2, weight: 0 }]);
		expect(result.nodeCount).toBe(2);
	});
});
