import { describe, expect, it } from "vitest";
import { communityColor, edgeElasticity, edgeWidth, idealEdgeLength, nodeSize } from "./graph";

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

describe("idealEdgeLength", () => {
	it("returns 150 for shared count of 1", () => {
		expect(idealEdgeLength(1)).toBeCloseTo(150);
	});

	it("returns 50 (floor) for high shared counts", () => {
		// 150 / sqrt(9) = 50, so shared >= 9 hits the floor
		expect(idealEdgeLength(9)).toBeCloseTo(50);
		expect(idealEdgeLength(100)).toBeCloseTo(50);
	});

	it("decreases as shared count increases", () => {
		expect(idealEdgeLength(1)).toBeGreaterThan(idealEdgeLength(4));
	});

	it("treats 0 the same as 1 (guards against division by zero)", () => {
		expect(idealEdgeLength(0)).toBeCloseTo(idealEdgeLength(1));
	});
});

describe("edgeElasticity", () => {
	it("returns 0.9 (cap) for high shared counts", () => {
		// 0.05 + 10/12 ≈ 0.88, so cap isn't hit until shared ≈ 10.2
		expect(edgeElasticity(11)).toBeCloseTo(0.9);
	});

	it("is monotonically increasing with shared count", () => {
		expect(edgeElasticity(1)).toBeLessThan(edgeElasticity(5));
	});

	it("treats 0 the same as 1 (guards against zero)", () => {
		expect(edgeElasticity(0)).toBeCloseTo(edgeElasticity(1));
	});

	it("never exceeds 0.9", () => {
		for (const shared of [1, 5, 10, 50, 100]) {
			expect(edgeElasticity(shared)).toBeLessThanOrEqual(0.9);
		}
	});
});
