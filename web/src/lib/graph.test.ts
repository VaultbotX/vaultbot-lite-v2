import { describe, expect, it } from "vitest";
import { communityColor, edgeWidth, nodeSize } from "./graph";

describe("communityColor", () => {
	it("returns a valid hsl string", () => {
		expect(communityColor(0, 4)).toMatch(/^hsl\(\d+, 62%, 56%\)$/);
	});

	it("is deterministic for the same inputs", () => {
		expect(communityColor(2, 8)).toBe(communityColor(2, 8));
	});

	it("returns different colors for different community IDs", () => {
		expect(communityColor(0, 4)).not.toBe(communityColor(1, 4));
	});

	it("keeps hue in [0, 360) for every community in a set", () => {
		for (let i = 0; i < 10; i++) {
			const color = communityColor(i, 10);
			const hue = Number(color.match(/hsl\((\d+)/)?.[1]);
			expect(hue).toBeGreaterThanOrEqual(0);
			expect(hue).toBeLessThan(360);
		}
	});

	it("offsets starting hue by ~200 for community 0", () => {
		// With numCommunities=1 the formula reduces to hsl(200, 62%, 56%)
		expect(communityColor(0, 1)).toBe("hsl(200, 62%, 56%)");
	});
});

describe("nodeSize", () => {
	it("returns the minimum size (18) when count is 0", () => {
		expect(nodeSize(0, 100)).toBeCloseTo(18);
	});

	it("returns the maximum size (58) when count equals maxCount", () => {
		expect(nodeSize(100, 100)).toBeCloseTo(58);
	});

	it("returns a value strictly between 18 and 58 for intermediate counts", () => {
		const size = nodeSize(50, 100);
		expect(size).toBeGreaterThan(18);
		expect(size).toBeLessThan(58);
	});

	it("is monotonically increasing with count", () => {
		const max = 100;
		expect(nodeSize(10, max)).toBeLessThan(nodeSize(30, max));
		expect(nodeSize(30, max)).toBeLessThan(nodeSize(70, max));
	});

	it("returns 18 when maxCount is 1 and count is 0", () => {
		// log(1) / log(2) = 1, log(0+1)/log(1+1) = 0 → 18 + 40*0 = 18
		expect(nodeSize(0, 1)).toBeCloseTo(18);
	});
});

describe("edgeWidth", () => {
	it("returns the minimum width (0.5) when count is 0", () => {
		expect(edgeWidth(0, 10)).toBeCloseTo(0.5);
	});

	it("returns the maximum width (4) when count equals maxShared", () => {
		expect(edgeWidth(10, 10)).toBeCloseTo(4);
	});

	it("returns a value strictly between 0.5 and 4 for intermediate counts", () => {
		const width = edgeWidth(5, 10);
		expect(width).toBeGreaterThan(0.5);
		expect(width).toBeLessThan(4);
	});

	it("is monotonically increasing with count", () => {
		expect(edgeWidth(2, 10)).toBeLessThan(edgeWidth(6, 10));
	});

	it("scales linearly: doubling count doubles the variable portion", () => {
		// variable portion = 3.5 * (count / maxShared)
		const base = 0.5;
		const w1 = edgeWidth(2, 10) - base;
		const w2 = edgeWidth(4, 10) - base;
		expect(w2).toBeCloseTo(w1 * 2);
	});
});
