import { describe, expect, it } from "vitest";
import {
	assignCommunityColors,
	COMMUNITY_PALETTE,
	desaturateColor,
	detailFetchUrl,
	edgeOpacity,
	edgeWidth,
	formatWindowRange,
	isolatedNodePosition,
	nodeSize,
	parseNodeParam,
	parseTimeRangeParams,
	rangesOverlap,
	type SearchableNode,
	searchNodes,
	type TimeRange,
} from "./graph";

describe("rangesOverlap", () => {
	it("returns false for an empty ranges array", () => {
		expect(rangesOverlap([], 0, 100)).toBe(false);
	});

	it("returns true when a range is fully inside the window", () => {
		expect(rangesOverlap([[10, 20]], 0, 100)).toBe(true);
	});

	it("returns true when the window is fully inside a range", () => {
		expect(rangesOverlap([[0, 100]], 10, 20)).toBe(true);
	});

	it("returns true when a range's start touches the window's end exactly", () => {
		expect(rangesOverlap([[100, 200]], 0, 100)).toBe(true);
	});

	it("returns true when a range's end touches the window's start exactly", () => {
		expect(rangesOverlap([[0, 100]], 100, 200)).toBe(true);
	});

	it("returns false for a range entirely before the window", () => {
		expect(rangesOverlap([[0, 50]], 51, 100)).toBe(false);
	});

	it("returns false for a range entirely after the window", () => {
		expect(rangesOverlap([[101, 150]], 0, 100)).toBe(false);
	});

	it("returns true when only one of several ranges overlaps", () => {
		const ranges: TimeRange[] = [
			[0, 10],
			[500, 600],
			[90, 110],
		];
		expect(rangesOverlap(ranges, 100, 200)).toBe(true);
	});

	it("returns false when none of several ranges overlap", () => {
		const ranges: TimeRange[] = [
			[0, 10],
			[500, 600],
		];
		expect(rangesOverlap(ranges, 100, 200)).toBe(false);
	});
});

describe("formatWindowRange", () => {
	it("formats start without a year and end with a year", () => {
		const start = new Date(2026, 5, 1).getTime() / 1000; // Jun 1, 2026 (local)
		const end = new Date(2026, 5, 15).getTime() / 1000; // Jun 15, 2026 (local)
		const result = formatWindowRange(start, end);
		expect(result).toContain("Jun 1");
		expect(result).toContain("Jun 15, 2026");
	});

	it("separates the two dates with an en dash", () => {
		const start = new Date(2026, 0, 1).getTime() / 1000;
		const end = new Date(2026, 0, 2).getTime() / 1000;
		expect(formatWindowRange(start, end)).toContain(" – ");
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

describe("edgeOpacity", () => {
	it("returns the minimum opacity (0.01) when count is 0", () => {
		expect(edgeOpacity(0, 10)).toBeCloseTo(0.01);
	});

	it("returns the maximum opacity (0.12) when count equals maxCount", () => {
		expect(edgeOpacity(10, 10)).toBeCloseTo(0.12);
	});

	it("is monotonically increasing with count", () => {
		expect(edgeOpacity(2, 10)).toBeLessThan(edgeOpacity(6, 10));
	});

	it("returns the minimum opacity when maxCount is 0, without dividing by zero", () => {
		expect(edgeOpacity(0, 0)).toBeCloseTo(0.01);
	});

	it("weights weak/mid edges much lower than a linear falloff would (quartic curve)", () => {
		// At half the max weight, opacity should sit far below the halfway point
		// between min and max opacity — that's the whole point of raising t to
		// the 4th power.
		const half = edgeOpacity(5, 10);
		const min = 0.01;
		const max = 0.12;
		expect(half).toBeLessThan(min + (max - min) / 2);
	});

	it("normalizes each edge kind against its own max rather than a shared max", () => {
		// A weight of 5 out of a small-scale max (artist-artist, max 5) should
		// read as "strong" even though the same raw weight would read as "weak"
		// against a large-scale max (genre-genre, max 500).
		const strongWithinOwnKind = edgeOpacity(5, 5);
		const weakAgainstOtherKindsMax = edgeOpacity(5, 500);
		expect(strongWithinOwnKind).toBeGreaterThan(weakAgainstOtherKindsMax);
	});
});

describe("desaturateColor", () => {
	it("returns the same color unchanged when amount is 0", () => {
		expect(desaturateColor("#DA654E", 0)).toBe("#da654e");
	});

	it("returns a pure gray (equal r/g/b) when amount is 1", () => {
		const result = desaturateColor("#DA654E", 1);
		const r = Number.parseInt(result.slice(1, 3), 16);
		const g = Number.parseInt(result.slice(3, 5), 16);
		const b = Number.parseInt(result.slice(5, 7), 16);
		expect(r).toBe(g);
		expect(g).toBe(b);
	});

	it("moves partway toward gray for an intermediate amount", () => {
		const full = desaturateColor("#DA654E", 1);
		const half = desaturateColor("#DA654E", 0.5);
		const grayValue = Number.parseInt(full.slice(1, 3), 16);
		const halfR = Number.parseInt(half.slice(1, 3), 16);
		const origR = 0xda;
		// Half-desaturated red channel should sit between the original and full gray.
		expect(halfR).toBeGreaterThan(Math.min(origR, grayValue));
		expect(halfR).toBeLessThan(Math.max(origR, grayValue));
	});

	it("preserves hue direction: a color already close to gray barely changes", () => {
		const nearGray = "#808080";
		expect(desaturateColor(nearGray, 0.5)).toBe("#808080");
	});
});

describe("isolatedNodePosition", () => {
	const centers = new Map([
		[0, { x: 100, y: 200 }],
		[1, { x: -50, y: 30 }],
	]);

	it("returns the community center when the community has a known center", () => {
		expect(isolatedNodePosition(0, centers)).toEqual({ x: 100, y: 200 });
		expect(isolatedNodePosition(1, centers)).toEqual({ x: -50, y: 30 });
	});

	it("falls back to the origin when the community has no known center", () => {
		expect(isolatedNodePosition(99, centers)).toEqual({ x: 0, y: 0 });
	});

	it("falls back to the origin when community is undefined", () => {
		expect(isolatedNodePosition(undefined, centers)).toEqual({ x: 0, y: 0 });
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

describe("searchNodes", () => {
	const nodes: SearchableNode[] = [
		{ id: 1, kind: "genre", name: "Pop" },
		{ id: 2, kind: "genre", name: "Pop Rock" },
		{ id: 3, kind: "artist", name: "Poppy" },
		{ id: 4, kind: "artist", name: "K-Pop Stars" },
		{ id: 5, kind: "genre", name: "Jazz" },
	];

	it("returns an empty array for an empty query", () => {
		expect(searchNodes("", nodes)).toEqual([]);
	});

	it("returns an empty array for a whitespace-only query", () => {
		expect(searchNodes("   ", nodes)).toEqual([]);
	});

	it("returns an empty array when nothing matches", () => {
		expect(searchNodes("xyz", nodes)).toEqual([]);
	});

	it("matches case-insensitively", () => {
		expect(searchNodes("JAZZ", nodes).map((n) => n.name)).toEqual(["Jazz"]);
	});

	it("ranks names that start with the query above names that only contain it", () => {
		const results = searchNodes("pop", nodes);
		// "Pop" and "Poppy" start with "pop"; "Pop Rock" also starts with it;
		// "K-Pop Stars" only contains it, so it must sort last.
		expect(results[results.length - 1].name).toBe("K-Pop Stars");
	});

	it("breaks ties among startsWith matches by shorter name first", () => {
		const results = searchNodes("pop", nodes);
		const startsWithNames = results
			.filter((n) => n.name.toLowerCase().startsWith("pop"))
			.map((n) => n.name);
		expect(startsWithNames).toEqual(["Pop", "Poppy", "Pop Rock"]);
	});

	it("respects the limit parameter", () => {
		expect(searchNodes("pop", nodes, 1).length).toBe(1);
	});

	it("defaults the limit to 8", () => {
		const many: SearchableNode[] = Array.from({ length: 10 }, (_, i) => ({
			id: i,
			kind: "genre",
			name: `Test ${i}`,
		}));
		expect(searchNodes("test", many).length).toBe(8);
	});
});

describe("parseTimeRangeParams", () => {
	it("returns null when both params are absent", () => {
		expect(parseTimeRangeParams(new URLSearchParams())).toBeNull();
	});

	it("returns null when only start is present", () => {
		expect(parseTimeRangeParams(new URLSearchParams("start=100"))).toBeNull();
	});

	it("returns null when only end is present", () => {
		expect(parseTimeRangeParams(new URLSearchParams("end=200"))).toBeNull();
	});

	it("parses a valid start/end pair", () => {
		expect(
			parseTimeRangeParams(new URLSearchParams("start=100&end=200")),
		).toEqual([100, 200]);
	});

	it("returns null when start is non-numeric", () => {
		expect(
			parseTimeRangeParams(new URLSearchParams("start=abc&end=200")),
		).toBeNull();
	});

	it("returns null when end is non-numeric", () => {
		expect(
			parseTimeRangeParams(new URLSearchParams("start=100&end=xyz")),
		).toBeNull();
	});

	it("returns null when start is after end", () => {
		expect(
			parseTimeRangeParams(new URLSearchParams("start=200&end=100")),
		).toBeNull();
	});

	it("accepts start equal to end", () => {
		expect(
			parseTimeRangeParams(new URLSearchParams("start=100&end=100")),
		).toEqual([100, 100]);
	});
});

describe("detailFetchUrl", () => {
	it("builds a bare genre URL when there is no active window", () => {
		expect(detailFetchUrl({ kind: "genre", id: 14 }, null)).toBe(
			"/api/genres/14",
		);
	});

	it("builds a bare artist URL when there is no active window", () => {
		expect(detailFetchUrl({ kind: "artist", id: 7 }, null)).toBe(
			"/api/artists/7",
		);
	});

	it("appends start/end params for a genre when a window is active", () => {
		expect(detailFetchUrl({ kind: "genre", id: 14 }, [100, 200])).toBe(
			"/api/genres/14?start=100&end=200",
		);
	});

	it("appends start/end params for an artist when a window is active", () => {
		expect(detailFetchUrl({ kind: "artist", id: 7 }, [100, 200])).toBe(
			"/api/artists/7?start=100&end=200",
		);
	});
});

describe("parseNodeParam", () => {
	it("parses a genre node param", () => {
		expect(parseNodeParam("g:14")).toEqual({ kind: "genre", id: 14 });
	});

	it("parses an artist node param", () => {
		expect(parseNodeParam("a:7")).toEqual({ kind: "artist", id: 7 });
	});

	it("returns null for null input", () => {
		expect(parseNodeParam(null)).toBeNull();
	});

	it("returns null for an empty string", () => {
		expect(parseNodeParam("")).toBeNull();
	});

	it("returns null for an unrecognized prefix", () => {
		expect(parseNodeParam("x:14")).toBeNull();
	});

	it("returns null for a non-numeric id", () => {
		expect(parseNodeParam("g:abc")).toBeNull();
	});

	it("returns null for a zero or negative id", () => {
		expect(parseNodeParam("g:0")).toBeNull();
		expect(parseNodeParam("a:-3")).toBeNull();
	});
});
