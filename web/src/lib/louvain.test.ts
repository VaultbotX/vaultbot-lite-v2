import { describe, expect, it } from "vitest";
import { detectCommunities } from "./louvain";

describe("detectCommunities", () => {
	it("returns empty map for empty input", () => {
		expect(detectCommunities([], []).size).toBe(0);
	});

	it("assigns each isolated node its own community when there are no edges", () => {
		const result = detectCommunities([1, 2, 3], []);
		expect(result.size).toBe(3);
		expect(new Set(result.values()).size).toBe(3);
	});

	it("places two directly connected nodes in the same community", () => {
		const result = detectCommunities(
			[1, 2],
			[{ source: 1, target: 2, weight: 1 }],
		);
		expect(result.get(1)).toBe(result.get(2));
	});

	it("merges a fully-connected triangle into one community", () => {
		const edges = [
			{ source: 1, target: 2, weight: 1 },
			{ source: 2, target: 3, weight: 1 },
			{ source: 1, target: 3, weight: 1 },
		];
		const result = detectCommunities([1, 2, 3], edges);
		expect(result.get(1)).toBe(result.get(2));
		expect(result.get(2)).toBe(result.get(3));
	});

	it("separates two dense triangles connected by a weak bridge", () => {
		const edges = [
			// cluster A
			{ source: 1, target: 2, weight: 10 },
			{ source: 2, target: 3, weight: 10 },
			{ source: 1, target: 3, weight: 10 },
			// cluster B
			{ source: 4, target: 5, weight: 10 },
			{ source: 5, target: 6, weight: 10 },
			{ source: 4, target: 6, weight: 10 },
			// weak bridge
			{ source: 3, target: 4, weight: 1 },
		];
		const result = detectCommunities([1, 2, 3, 4, 5, 6], edges);
		expect(result.get(1)).toBe(result.get(2));
		expect(result.get(2)).toBe(result.get(3));
		expect(result.get(4)).toBe(result.get(5));
		expect(result.get(5)).toBe(result.get(6));
		expect(result.get(1)).not.toBe(result.get(4));
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
		const ids = [...new Set(result.values())].sort((a, b) => a - b);
		expect(ids[0]).toBe(0);
		for (let i = 1; i < ids.length; i++) {
			expect(ids[i]).toBe(ids[i - 1] + 1);
		}
	});

	it("ignores self-loop edges", () => {
		// Self-loop (source === target) must not crash and is skipped
		const result = detectCommunities(
			[1, 2],
			[{ source: 1, target: 1, weight: 5 }],
		);
		expect(result.size).toBe(2);
	});

	it("handles nodes not referenced by any edge", () => {
		const edges = [{ source: 1, target: 2, weight: 5 }];
		// node 3 is isolated — no edges touch it
		const result = detectCommunities([1, 2, 3], edges);
		expect(result.size).toBe(3);
		expect(result.get(1)).toBe(result.get(2));
		// node 3 should be in its own community
		expect(result.get(3)).not.toBe(result.get(1));
	});

	it("handles edges with zero weight without crashing", () => {
		const result = detectCommunities(
			[1, 2],
			[{ source: 1, target: 2, weight: 0 }],
		);
		expect(result.size).toBe(2);
	});
});
