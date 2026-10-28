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
	assignCommunityColors,
	COMMUNITY_PALETTE,
	desaturateColor,
	edgeOpacity,
	edgePairKey,
	edgeWidth,
} from "./graph";

export type NodeKind = "genre" | "artist";

// Edges share one color; their per-edge fade is the native `opacity` style.
const EDGE_COLOR = "#6060a0";

function genreNodeId(genreId: number): string {
	return `g:${genreId}`;
}

function artistNodeId(artistId: number): string {
	return `a:${artistId}`;
}

// Quadratic scale on the log-ratio: small nodes collapse toward ~1.5,
// large nodes grow toward 20. The squaring amplifies the size gap
// compared to a plain linear mapping.
function scaledNodeSize(count: number, maxCount: number): number {
	const t = Math.log(count + 1) / Math.log(maxCount + 1);
	return 1.5 + 18.5 * t * t;
}

/**
 * Builds a fully-decorated graphology Graph mixing genre and artist nodes,
 * with genre-genre, genre-artist, and artist-artist edges.
 *
 * Node attributes set: label, kind, genreId | artistId, size, color, community
 * Edge attributes set: kind, weight, size (width), color, opacity
 *
 * Each node/edge kind is size-normalized against its own kind's max, not a
 * global max, since genre archive counts (summed across many artists) live
 * on a much larger scale than a single artist's archive count.
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
			label: `${v.name} (#${v.rank})`,
			kind: "genre" satisfies NodeKind,
			genreId: v.genre_id,
			archiveCount: v.archive_count,
		});
	}

	for (const v of artistVertices) {
		graph.addNode(artistNodeId(v.artist_id), {
			label: `${v.name} (#${v.rank})`,
			kind: "artist" satisfies NodeKind,
			artistId: v.artist_id,
			archiveCount: v.archive_count,
		});
	}

	const addWeightedEdge = (
		src: string,
		tgt: string,
		kind: "genre-genre" | "genre-artist" | "artist-artist",
		weight: number,
		maxWeight: number,
		extraAttrs: { pairKey?: string; artistId?: number },
	): void => {
		if (!graph.hasNode(src) || !graph.hasNode(tgt) || graph.hasEdge(src, tgt))
			return;
		const opacity = edgeOpacity(weight, maxWeight);
		graph.addEdge(src, tgt, {
			kind,
			weight,
			size: edgeWidth(weight, maxWeight),
			color: EDGE_COLOR,
			opacity,
			...extraAttrs,
		});
	};

	for (const e of genreGenreEdges) {
		addWeightedEdge(
			genreNodeId(e.source_genre_id),
			genreNodeId(e.target_genre_id),
			"genre-genre",
			e.shared_archive_count,
			maxGenreGenreWeight,
			{ pairKey: edgePairKey(e.source_genre_id, e.target_genre_id) },
		);
	}

	for (const e of genreArtistEdges) {
		const weight = artistArchiveById.get(e.artist_id) ?? 0;
		if (weight <= 0) continue;
		addWeightedEdge(
			genreNodeId(e.genre_id),
			artistNodeId(e.artist_id),
			"genre-artist",
			weight,
			maxGenreArtistWeight,
			{ artistId: e.artist_id },
		);
	}

	for (const e of artistArtistEdges) {
		addWeightedEdge(
			artistNodeId(e.source_artist_id),
			artistNodeId(e.target_artist_id),
			"artist-artist",
			e.shared_song_count,
			maxArtistArtistWeight,
			{ pairKey: edgePairKey(e.source_artist_id, e.target_artist_id) },
		);
	}

	// Detect communities over the full mixed graph, weighted by raw edge weight.
	louvain.assign(graph, { getEdgeWeight: "weight" });

	// Assign palette colors by community ID
	const communityIds = new Set<number>();
	graph.forEachNode((_, attrs) => communityIds.add(attrs.community as number));
	const colorMap = assignCommunityColors(communityIds, COMMUNITY_PALETTE);

	// Artist fills are desaturated relative to their genre counterparts within
	// the same community — same hue, muted intensity — so the two node kinds
	// stay visually distinguishable even in a dense, same-colored cluster.
	const ARTIST_DESATURATION = 0.35;

	graph.forEachNode((node, attrs) => {
		const kind = attrs.kind as NodeKind;
		const archiveCount = attrs.archiveCount as number;
		const maxForKind = kind === "genre" ? maxGenreArchive : maxArtistArchive;
		const size = scaledNodeSize(archiveCount, maxForKind);
		const baseColor = colorMap.get(attrs.community as number) ?? "#888888";

		graph.setNodeAttribute(node, "size", size);
		graph.setNodeAttribute(
			node,
			"color",
			kind === "artist"
				? desaturateColor(baseColor, ARTIST_DESATURATION)
				: baseColor,
		);
	});

	return graph;
}
