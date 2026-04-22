import { describe, expect, it } from "vitest";
import { Community, GenreEdge, GenreGraph, GenreNode } from "./genre-graph";
import { CommunityPartition } from "./louvain";
import { COMMUNITY_PALETTE } from "./graph";

// ── helpers ─────────────────────────────────────────────────────────────────

function makeComm(id = 0, color = "red"): Community {
	return new Community(id, color, new Set([id * 10 + 1]));
}

function makeNode(
	genreId = 1,
	name = "genre",
	artistCount = 5,
	community = makeComm(),
): GenreNode {
	return new GenreNode(genreId, name, artistCount, community);
}

// ── Community ────────────────────────────────────────────────────────────────

describe("Community", () => {
	it("exposes id, color, and memberIds", () => {
		const c = new Community(3, "blue", new Set([1, 2]));
		expect(c.id).toBe(3);
		expect(c.color).toBe("blue");
		expect(c.memberIds).toEqual(new Set([1, 2]));
	});

	it("memberCount equals the number of members", () => {
		expect(new Community(0, "red", new Set([1, 2, 3])).memberCount).toBe(3);
	});
});

// ── GenreNode ────────────────────────────────────────────────────────────────

describe("GenreNode", () => {
	it("displayColor delegates to its community color", () => {
		const node = makeNode(1, "pop", 10, new Community(0, "teal", new Set([1])));
		expect(node.displayColor).toBe("teal");
	});

	it("displaySize returns minimum when artistCount is 0", () => {
		const node = makeNode(1, "pop", 0);
		expect(node.displaySize(100)).toBeCloseTo(14);
	});

	it("displaySize returns maximum when artistCount equals maxArtistCount", () => {
		const node = makeNode(1, "pop", 100);
		expect(node.displaySize(100)).toBeCloseTo(64);
	});

	it("displaySize is monotonically increasing", () => {
		const max = 100;
		expect(makeNode(1, "a", 10).displaySize(max)).toBeLessThan(
			makeNode(1, "b", 50).displaySize(max),
		);
	});
});

// ── GenreEdge ────────────────────────────────────────────────────────────────

describe("GenreEdge", () => {
	const src = makeNode(1, "a", 5);
	const tgt = makeNode(2, "b", 5);

	it("displayWidth returns minimum (0.5) when sharedArtistCount is 0", () => {
		const edge = new GenreEdge(src, tgt, 0);
		expect(edge.displayWidth(10)).toBeCloseTo(0.5);
	});

	it("displayWidth returns maximum (5.5) when sharedArtistCount equals maxShared", () => {
		const edge = new GenreEdge(src, tgt, 10);
		expect(edge.displayWidth(10)).toBeCloseTo(5.5);
	});

	it("displayOpacity is in (0, 1] range", () => {
		for (const shared of [0, 1, 5, 10]) {
			const opacity = new GenreEdge(src, tgt, shared).displayOpacity(10);
			expect(opacity).toBeGreaterThan(0);
			expect(opacity).toBeLessThanOrEqual(1);
		}
	});

	it("displayOpacity increases with sharedArtistCount", () => {
		const low = new GenreEdge(src, tgt, 1).displayOpacity(10);
		const high = new GenreEdge(src, tgt, 9).displayOpacity(10);
		expect(low).toBeLessThan(high);
	});
});

// ── GenreGraph.build ─────────────────────────────────────────────────────────

describe("GenreGraph.build", () => {
	const vertices = [
		{ genre_id: 1, name: "pop", artist_count: 10 },
		{ genre_id: 2, name: "rock", artist_count: 5 },
		{ genre_id: 3, name: "jazz", artist_count: 3 },
	];
	const apiEdges = [
		{ source_genre_id: 1, target_genre_id: 2, shared_artist_count: 3 },
		{ source_genre_id: 2, target_genre_id: 3, shared_artist_count: 1 },
	];
	// Put 1 and 2 in community 0, 3 in community 1
	const partition = new CommunityPartition(new Map([[1, 0], [2, 0], [3, 1]]));

	it("builds a graph with the correct node count", () => {
		const g = GenreGraph.build(vertices, apiEdges, partition);
		expect(g.nodes.length).toBe(3);
	});

	it("builds a graph with the correct edge count", () => {
		const g = GenreGraph.build(vertices, apiEdges, partition);
		expect(g.edges.length).toBe(2);
	});

	it("builds the correct number of communities", () => {
		const g = GenreGraph.build(vertices, apiEdges, partition);
		expect(g.communities.size).toBe(2);
	});

	it("nodes in the same community share a color", () => {
		const g = GenreGraph.build(vertices, apiEdges, partition);
		const pop = g.nodes.find((n) => n.name === "pop")!;
		const rock = g.nodes.find((n) => n.name === "rock")!;
		expect(pop.displayColor).toBe(rock.displayColor);
	});

	it("nodes in different communities have different colors", () => {
		const g = GenreGraph.build(vertices, apiEdges, partition);
		const pop = g.nodes.find((n) => n.name === "pop")!;
		const jazz = g.nodes.find((n) => n.name === "jazz")!;
		expect(pop.displayColor).not.toBe(jazz.displayColor);
	});

	it("skips edges whose endpoints are missing from vertices", () => {
		const orphanEdges = [
			{ source_genre_id: 1, target_genre_id: 99, shared_artist_count: 5 },
		];
		const g = GenreGraph.build(vertices, orphanEdges, partition);
		expect(g.edges.length).toBe(0);
	});

	it("uses the provided palette for community colors", () => {
		const palette = ["#aaa", "#bbb"];
		const g = GenreGraph.build(vertices, apiEdges, partition, palette);
		for (const node of g.nodes) {
			expect(palette).toContain(node.displayColor);
		}
	});
});

// ── GenreGraph.nodeDisplays / edgeDisplays ───────────────────────────────────

describe("GenreGraph.nodeDisplays", () => {
	const partition = new CommunityPartition(new Map([[1, 0], [2, 0]]));
	const vertices = [
		{ genre_id: 1, name: "pop", artist_count: 10 },
		{ genre_id: 2, name: "rock", artist_count: 5 },
	];
	const g = GenreGraph.build(vertices, [], partition);

	it("returns one NodeDisplay per node", () => {
		expect(g.nodeDisplays().length).toBe(2);
	});

	it("NodeDisplay id is the string form of genreId", () => {
		const display = g.nodeDisplays().find((d) => d.genreId === 1)!;
		expect(display.id).toBe("1");
	});

	it("NodeDisplay label matches vertex name", () => {
		const display = g.nodeDisplays().find((d) => d.genreId === 1)!;
		expect(display.label).toBe("pop");
	});

	it("NodeDisplay size is within expected range", () => {
		for (const d of g.nodeDisplays()) {
			expect(d.size).toBeGreaterThanOrEqual(14);
			expect(d.size).toBeLessThanOrEqual(64);
		}
	});
});

describe("GenreGraph.edgeDisplays", () => {
	const partition = new CommunityPartition(new Map([[1, 0], [2, 1]]));
	const vertices = [
		{ genre_id: 1, name: "pop", artist_count: 10 },
		{ genre_id: 2, name: "rock", artist_count: 5 },
	];
	const apiEdges = [{ source_genre_id: 1, target_genre_id: 2, shared_artist_count: 3 }];
	const g = GenreGraph.build(vertices, apiEdges, partition);

	it("returns one EdgeDisplay per edge", () => {
		expect(g.edgeDisplays().length).toBe(1);
	});

	it("EdgeDisplay sourceId and targetId are string genre IDs", () => {
		const [d] = g.edgeDisplays();
		expect(d.sourceId).toBe("1");
		expect(d.targetId).toBe("2");
	});

	it("EdgeDisplay width is positive", () => {
		const [d] = g.edgeDisplays();
		expect(d.width).toBeGreaterThan(0);
	});

	it("EdgeDisplay opacity is in (0, 1]", () => {
		const [d] = g.edgeDisplays();
		expect(d.opacity).toBeGreaterThan(0);
		expect(d.opacity).toBeLessThanOrEqual(1);
	});
});
