import Graph from "graphology";
import louvain from "graphology-communities-louvain";
import type {
	ArtistArtistEdge,
	ArtistVertex,
	GenreArtistEdge,
	GenreGenreEdge,
	GenreVertex,
} from "../routes/api/graph/+server";
import {
	assignCommunityColorsBySize,
	COMMUNITY_PALETTE,
	createSeededRng,
	desaturateColor,
	edgePairKey,
	logNormalizedWeight,
	selectBackboneEdges,
} from "./graph";

export type NodeKind = "genre" | "artist";

const INTER_COMMUNITY_EDGE_COLOR = "#505078";
const ISOLATED_NODE_COLOR = "#52526e";
const LOUVAIN_SEED = 42;
const LOUVAIN_RESOLUTION = 1.25;
const GENRE_GENRE_TOP_K = 4;
const GENRE_ARTIST_TOP_K = 3;
const ARTIST_ARTIST_TOP_K = 3;

function genreNodeId(genreId: number): string {
	return `g:${genreId}`;
}

function artistNodeId(artistId: number): string {
	return `a:${artistId}`;
}

// Genres serve as larger constellation hubs; artists orbit them as smaller stars.
function scaledNodeSize(
	count: number,
	maxCount: number,
	kind: NodeKind,
): number {
	const t = Math.log(count + 1) / Math.log(maxCount + 1);
	return kind === "genre" ? 2.2 + 13.8 * t * t : 1.4 + 8.6 * t * t;
}

/**
 * Builds a fully-decorated graphology Graph mixing genre and artist nodes,
 * with genre-genre, genre-artist, and artist-artist edges.
 *
 * Node attributes set: label, rank, kind, genreId | artistId, size, color, community
 * Edge attributes set: kind, weight, normWeight, layoutWeight, backbone, size, color, opacity
 *
 * Positions (x, y) are NOT set here — the layout algorithm in the
 * rendering component assigns those before handing off to sigma.
 */
export function buildMixedGraph(
	genreVertices: GenreVertex[],
	artistVertices: ArtistVertex[],
	genreGenreEdges: GenreGenreEdge[],
	genreArtistEdges: GenreArtistEdge[],
	artistArtistEdges: ArtistArtistEdge[],
): Graph {
	const graph = new Graph({ type: "undirected", multi: false });

	if (genreVertices.length === 0 && artistVertices.length === 0) return graph;

	const artistArchiveById = new Map<number, number>();
	let maxGenreArchive = 1;
	for (const v of genreVertices) {
		if (v.archive_count > maxGenreArchive) maxGenreArchive = v.archive_count;
	}
	let maxArtistArchive = 1;
	for (const v of artistVertices) {
		artistArchiveById.set(v.artist_id, v.archive_count);
		if (v.archive_count > maxArtistArchive) maxArtistArchive = v.archive_count;
	}
	let maxGenreGenreWeight = 1;
	for (const e of genreGenreEdges) {
		if (e.shared_archive_count > maxGenreGenreWeight)
			maxGenreGenreWeight = e.shared_archive_count;
	}
	let maxGenreArtistWeight = 1;
	for (const e of genreArtistEdges) {
		const count = artistArchiveById.get(e.artist_id) ?? 0;
		if (count > maxGenreArtistWeight) maxGenreArtistWeight = count;
	}
	let maxArtistArtistWeight = 1;
	for (const e of artistArtistEdges) {
		if (e.shared_song_count > maxArtistArtistWeight)
			maxArtistArtistWeight = e.shared_song_count;
	}

	for (const v of genreVertices) {
		graph.addNode(genreNodeId(v.genre_id), {
			label: v.name,
			rank: v.rank,
			kind: "genre" satisfies NodeKind,
			genreId: v.genre_id,
			archiveCount: v.archive_count,
		});
	}

	for (const v of artistVertices) {
		graph.addNode(artistNodeId(v.artist_id), {
			label: v.name,
			rank: v.rank,
			kind: "artist" satisfies NodeKind,
			artistId: v.artist_id,
			archiveCount: v.archive_count,
		});
	}

	const validGenreArtistEdges = genreArtistEdges.filter(
		(e) => (artistArchiveById.get(e.artist_id) ?? 0) > 0,
	);

	const ggBackbone = selectBackboneEdges(
		genreGenreEdges,
		(e) => genreNodeId(e.source_genre_id),
		(e) => genreNodeId(e.target_genre_id),
		(e) => e.shared_archive_count,
		GENRE_GENRE_TOP_K,
	);
	const gaBackbone = selectBackboneEdges(
		validGenreArtistEdges,
		(e) => genreNodeId(e.genre_id),
		(e) => artistNodeId(e.artist_id),
		(e) => artistArchiveById.get(e.artist_id) ?? 0,
		GENRE_ARTIST_TOP_K,
	);
	const aaBackbone = selectBackboneEdges(
		artistArtistEdges,
		(e) => artistNodeId(e.source_artist_id),
		(e) => artistNodeId(e.target_artist_id),
		(e) => e.shared_song_count,
		ARTIST_ARTIST_TOP_K,
	);

	// Subgraph containing only connected nodes and backbone edges so Louvain
	// clusters cleanly without hairball cross-links or degree-0 singletons.
	const backboneGraph = new Graph({ type: "undirected", multi: false });

	const addWeightedEdge = (
		src: string,
		tgt: string,
		kind: "genre-genre" | "genre-artist" | "artist-artist",
		weight: number,
		maxWeight: number,
		backbone: boolean,
		extraAttrs: { pairKey?: string; artistId?: number },
	): void => {
		if (!graph.hasNode(src) || !graph.hasNode(tgt) || graph.hasEdge(src, tgt))
			return;
		const normWeight = logNormalizedWeight(weight, maxWeight);
		graph.addEdge(src, tgt, {
			kind,
			weight,
			normWeight,
			backbone,
			...extraAttrs,
		});
		if (backbone) {
			if (!backboneGraph.hasNode(src)) backboneGraph.addNode(src);
			if (!backboneGraph.hasNode(tgt)) backboneGraph.addNode(tgt);
			if (!backboneGraph.hasEdge(src, tgt)) {
				backboneGraph.addEdge(src, tgt, { weight: normWeight });
			}
		}
	};

	for (let i = 0; i < genreGenreEdges.length; i++) {
		const e = genreGenreEdges[i];
		addWeightedEdge(
			genreNodeId(e.source_genre_id),
			genreNodeId(e.target_genre_id),
			"genre-genre",
			e.shared_archive_count,
			maxGenreGenreWeight,
			ggBackbone.has(i),
			{ pairKey: edgePairKey(e.source_genre_id, e.target_genre_id) },
		);
	}

	for (let i = 0; i < validGenreArtistEdges.length; i++) {
		const e = validGenreArtistEdges[i];
		const weight = artistArchiveById.get(e.artist_id) ?? 0;
		addWeightedEdge(
			genreNodeId(e.genre_id),
			artistNodeId(e.artist_id),
			"genre-artist",
			weight,
			maxGenreArtistWeight,
			gaBackbone.has(i),
			{ artistId: e.artist_id },
		);
	}

	for (let i = 0; i < artistArtistEdges.length; i++) {
		const e = artistArtistEdges[i];
		addWeightedEdge(
			artistNodeId(e.source_artist_id),
			artistNodeId(e.target_artist_id),
			"artist-artist",
			e.shared_song_count,
			maxArtistArtistWeight,
			aaBackbone.has(i),
			{ pairKey: edgePairKey(e.source_artist_id, e.target_artist_id) },
		);
	}

	if (backboneGraph.order > 0) {
		louvain.assign(backboneGraph, {
			getEdgeWeight: "weight",
			resolution: LOUVAIN_RESOLUTION,
			rng: createSeededRng(LOUVAIN_SEED),
		});
	}

	const communitySizes = new Map<number, number>();
	backboneGraph.forEachNode((node, attrs) => {
		const comm = attrs.community as number;
		graph.setNodeAttribute(node, "community", comm);
		communitySizes.set(comm, (communitySizes.get(comm) ?? 0) + 1);
	});

	const colorMap = assignCommunityColorsBySize(
		communitySizes,
		COMMUNITY_PALETTE,
	);

	// Artist fills are desaturated relative to their genre counterparts within
	// the same community — same hue, muted intensity — so the two node kinds
	// stay visually distinguishable even in a dense, same-colored cluster.
	const ARTIST_DESATURATION = 0.32;

	graph.forEachNode((node, attrs) => {
		const kind = attrs.kind as NodeKind;
		const archiveCount = attrs.archiveCount as number;
		const maxForKind = kind === "genre" ? maxGenreArchive : maxArtistArchive;
		const size = scaledNodeSize(archiveCount, maxForKind, kind);
		const comm = attrs.community as number | undefined;

		graph.setNodeAttribute(node, "size", size);
		if (comm === undefined) {
			graph.setNodeAttribute(node, "community", -1);
			graph.setNodeAttribute(node, "color", ISOLATED_NODE_COLOR);
			return;
		}
		const baseColor = colorMap.get(comm) ?? ISOLATED_NODE_COLOR;
		graph.setNodeAttribute(
			node,
			"color",
			kind === "artist"
				? desaturateColor(baseColor, ARTIST_DESATURATION)
				: baseColor,
		);
	});

	// Decorate edges now that community membership and colors are known:
	// intra-community backbone filaments inherit a muted community tint and full
	// layout cohesion, whereas cross-community bridges stay faint and exert
	// weaker spring attraction so distinct constellations separate cleanly.
	graph.forEachEdge((edge, attrs, src, tgt) => {
		const normWeight = attrs.normWeight as number;
		const backbone = attrs.backbone as boolean;
		const srcComm = graph.getNodeAttribute(src, "community") as number;
		const tgtComm = graph.getNodeAttribute(tgt, "community") as number;
		const sameCommunity = srcComm >= 0 && srcComm === tgtComm;

		if (!backbone) {
			graph.setEdgeAttribute(edge, "color", INTER_COMMUNITY_EDGE_COLOR);
			graph.setEdgeAttribute(edge, "opacity", 0.02);
			graph.setEdgeAttribute(edge, "size", 0.4 + 1.2 * normWeight);
			graph.setEdgeAttribute(edge, "layoutWeight", 0);
			return;
		}

		if (sameCommunity) {
			const baseColor = colorMap.get(srcComm) ?? INTER_COMMUNITY_EDGE_COLOR;
			graph.setEdgeAttribute(edge, "color", desaturateColor(baseColor, 0.45));
			graph.setEdgeAttribute(
				edge,
				"opacity",
				0.04 + 0.24 * normWeight * normWeight,
			);
			graph.setEdgeAttribute(edge, "size", 0.45 + 2.0 * normWeight);
			graph.setEdgeAttribute(edge, "layoutWeight", normWeight);
		} else {
			graph.setEdgeAttribute(edge, "color", INTER_COMMUNITY_EDGE_COLOR);
			graph.setEdgeAttribute(
				edge,
				"opacity",
				0.015 + 0.075 * normWeight * normWeight,
			);
			graph.setEdgeAttribute(edge, "size", 0.4 + 1.1 * normWeight);
			graph.setEdgeAttribute(edge, "layoutWeight", normWeight * 0.12);
		}
	});

	return graph;
}
