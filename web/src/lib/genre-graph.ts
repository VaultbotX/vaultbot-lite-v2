import Graph from "graphology";
import louvain from "graphology-communities-louvain";
import type { GenreEdge, GenreVertex } from "../routes/api/graph/+server";
import {
	assignCommunityColors,
	COMMUNITY_PALETTE,
	edgeWidth,
} from "./graph";

/**
 * Builds a fully-decorated graphology Graph from raw API data.
 *
 * Node attributes set: label, genreId, size, color, community
 * Edge attributes set: shared, size (width), color (rgba string)
 *
 * Positions (x, y) are NOT set here — the layout algorithm in the
 * rendering component assigns those before handing off to sigma.
 */
export function buildGenreGraph(
	vertices: GenreVertex[],
	edges: GenreEdge[],
): Graph {
	const graph = new Graph({ type: "undirected", multi: false });

	if (vertices.length === 0) return graph;

	const maxArtistCount = Math.max(...vertices.map((v) => v.artist_count), 1);
	const maxShared = Math.max(...edges.map((e) => e.shared_artist_count), 1);

	for (const v of vertices) {
		graph.addNode(String(v.genre_id), {
			label: v.name,
			genreId: v.genre_id,
			artistCount: v.artist_count,
		});
	}

	for (const e of edges) {
		const src = String(e.source_genre_id);
		const tgt = String(e.target_genre_id);
		if (graph.hasNode(src) && graph.hasNode(tgt) && !graph.hasEdge(src, tgt)) {
			const opacity = 0.15 + 0.5 * Math.sqrt(e.shared_artist_count / maxShared);
			graph.addEdge(src, tgt, {
				shared: e.shared_artist_count,
				size: edgeWidth(e.shared_artist_count, maxShared),
				color: `rgba(96, 96, 160, ${opacity.toFixed(2)})`,
			});
		}
	}

	// Detect communities weighted by shared artist count
	louvain.assign(graph, { getEdgeWeight: "shared" });

	// Assign palette colors by community ID
	const communityIds = new Set<number>();
	graph.forEachNode((_, attrs) => communityIds.add(attrs.community as number));
	const colorMap = assignCommunityColors(communityIds, COMMUNITY_PALETTE);

	graph.forEachNode((node, attrs) => {
		const artistCount = attrs.artistCount as number;
		// Quadratic scale on the log-ratio: small nodes collapse toward ~1.5,
		// large nodes grow toward 20. The squaring amplifies the size gap
		// compared to a plain linear mapping.
		const t = Math.log(artistCount + 1) / Math.log(maxArtistCount + 1);
		const size = 1.5 + 18.5 * t * t;

		graph.setNodeAttribute(node, "size", size);
		graph.setNodeAttribute(
			node,
			"color",
			colorMap.get(attrs.community as number) ?? "#888888",
		);
	});

	return graph;
}
