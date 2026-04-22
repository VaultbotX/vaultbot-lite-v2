import { describe, expect, it } from "vitest";
import {
	assignCommunityColors,
	COMMUNITY_PALETTE,
	edgeWidth,
	nodeSize,
} from "./graph";

describe("nodeSize", () => {
	it("returns the minimum size (14) when count is 0", () => {
		expect(nodeSize(0, 100)).toBeCloseTo(14);
	});

	it("returns the maximum size (64) when count equals maxCount", () => {
		expect(nodeSize(100, 100)).toBeCloseTo(64);
	});

	it("returns a value strictly between 14 and 64 for intermediate counts", () => {
		const size = nodeSize(50, 100);
		expect(size).toBeGreaterThan(14);
		expect(size).toBeLessThan(64);
	});

	it("is monotonically increasing with count", () => {
		const max = 100;
		expect(nodeSize(10, max)).toBeLessThan(nodeSize(30, max));
		expect(nodeSize(30, max)).toBeLessThan(nodeSize(70, max));
	});

	it("returns 14 when maxCount is 1 and count is 0", () => {
		// log(0+1)/log(1+1) = 0 → 14 + 50*0 = 14
		expect(nodeSize(0, 1)).toBeCloseTo(14);
	});
});

describe("edgeWidth", () => {
	it("returns the minimum width (0.5) when count is 0", () => {
		expect(edgeWidth(0, 10)).toBeCloseTo(0.5);
	});

	it("returns the maximum width (5.5) when count equals maxShared", () => {
		expect(edgeWidth(10, 10)).toBeCloseTo(5.5);
	});

	it("returns a value strictly between 0.5 and 5.5 for intermediate counts", () => {
		const width = edgeWidth(5, 10);
		expect(width).toBeGreaterThan(0.5);
		expect(width).toBeLessThan(5.5);
	});

	it("is monotonically increasing with count", () => {
		expect(edgeWidth(2, 10)).toBeLessThan(edgeWidth(6, 10));
	});

	it("scales by sqrt: variable portion grows with square root of count ratio", () => {
		// variable portion = 5 * sqrt(count / maxShared)
		const base = 0.5;
		const w1 = edgeWidth(1, 10) - base;
		const w4 = edgeWidth(4, 10) - base;
		// sqrt(4/10) / sqrt(1/10) = sqrt(4) = 2
		expect(w4).toBeCloseTo(w1 * 2);
	});
});

const P = ["red", "blue", "green", "yellow", "orange"];

describe("assignCommunityColors", () => {
	it("returns an empty map when given no community IDs", () => {
		expect(assignCommunityColors([], P).size).toBe(0);
	});

	it("assigns a color to a single community", () => {
		const result = assignCommunityColors([0], P);
		expect(result.size).toBe(1);
		expect(result.get(0)).toBeDefined();
	});

	it("assigns different colors to different community IDs", () => {
		const result = assignCommunityColors([0, 1, 2], P);
		expect(result.get(0)).not.toBe(result.get(1));
		expect(result.get(1)).not.toBe(result.get(2));
		expect(result.get(0)).not.toBe(result.get(2));
	});

	it("result contains one entry per unique community ID", () => {
		// Duplicates in the iterable are deduped
		const result = assignCommunityColors([0, 0, 1, 1], P);
		expect(result.size).toBe(2);
	});

	it("all assigned colors come from the palette", () => {
		const result = assignCommunityColors([0, 1, 2], P);
		for (const color of result.values()) {
			expect(P).toContain(color);
		}
	});

	it("cycles through palette when community count exceeds palette length", () => {
		// 6 communities, 5 palette entries → commId 5 wraps to palette[0]
		const result = assignCommunityColors([0, 1, 2, 3, 4, 5], P);
		expect(result.get(5)).toBe(result.get(0));
	});

	it("assignment is deterministic for the same input", () => {
		const r1 = assignCommunityColors([0, 2, 1], P);
		const r2 = assignCommunityColors([0, 2, 1], P);
		for (const [id, color] of r1) {
			expect(r2.get(id)).toBe(color);
		}
	});

	it("sorts by community ID so palette[0] always goes to the lowest ID", () => {
		const result = assignCommunityColors([3, 1, 0], P);
		expect(result.get(0)).toBe(P[0]);
		expect(result.get(1)).toBe(P[1]);
		expect(result.get(3)).toBe(P[2]);
	});

	it("COMMUNITY_PALETTE has 12 entries so up to 12 communities get unique colors", () => {
		const ids = Array.from({ length: 12 }, (_, i) => i);
		const result = assignCommunityColors(ids, COMMUNITY_PALETTE);
		expect(new Set(result.values()).size).toBe(12);
	});
});
