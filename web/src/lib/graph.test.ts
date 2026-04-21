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
	it("returns the minimum size (14) when count is 0", () => {
		expect(nodeSize(0, 100)).toBeCloseTo(14);
	});

	it("returns the maximum size (112) when count equals maxCount", () => {
		expect(nodeSize(100, 100)).toBeCloseTo(112);
	});

	it("returns a value strictly between 14 and 112 for intermediate counts", () => {
		const size = nodeSize(50, 100);
		expect(size).toBeGreaterThan(14);
		expect(size).toBeLessThan(112);
	});

	it("is monotonically increasing with count", () => {
		const max = 100;
		expect(nodeSize(10, max)).toBeLessThan(nodeSize(30, max));
		expect(nodeSize(30, max)).toBeLessThan(nodeSize(70, max));
	});

	it("returns 14 when maxCount is 1 and count is 0", () => {
		// log(0+1)/log(1+1) = 0 → 14 + 98*0 = 14
		expect(nodeSize(0, 1)).toBeCloseTo(14);
	});
});

describe("edgeWidth", () => {
	it("returns the minimum width (0.5) when count is 0", () => {
		expect(edgeWidth(0, 10)).toBeCloseTo(0.5);
	});

	it("returns the maximum width (14) when count equals maxShared", () => {
		expect(edgeWidth(10, 10)).toBeCloseTo(14);
	});

	it("returns a value strictly between 0.5 and 14 for intermediate counts", () => {
		const width = edgeWidth(5, 10);
		expect(width).toBeGreaterThan(0.5);
		expect(width).toBeLessThan(14);
	});

	it("is monotonically increasing with count", () => {
		expect(edgeWidth(2, 10)).toBeLessThan(edgeWidth(6, 10));
	});

	it("scales by sqrt: variable portion grows with square root of count ratio", () => {
		// variable portion = 13.5 * sqrt(count / maxShared)
		const base = 0.5;
		const w1 = edgeWidth(1, 10) - base;
		const w4 = edgeWidth(4, 10) - base;
		// sqrt(4/10) / sqrt(1/10) = sqrt(4) = 2
		expect(w4).toBeCloseTo(w1 * 2);
	});
});
