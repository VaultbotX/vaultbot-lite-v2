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
	edgeOpacity,
	edgeWidth,
} from "./graph";

export type NodeKind = "genre" | "artist";

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
 * Edge attributes set: kind, weight, size (width), color (rgba string)
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

	const maxGenreArchive = Math.max(
		...genreVertices.map((v) => v.archive_count),
		1,
	);
	const maxArtistArchive = Math.max(
		...artistVertices.map((v) => v.archive_count),
		1,
	);
	const maxGenreGenreWeight = Math.max(
		...genreGenreEdges.map((e) => e.shared_archive_count),
		1,
	);
	const maxGenreArtistWeight = Math.max(
		...genreArtistEdges.map((e) => e.archive_count),
		1,
	);
	const maxArtistArtistWeight = Math.max(
		...artistArtistEdges.map((e) => e.shared_song_count),
		1,
	);

	// Border color doubles as the genre/artist visual disambiguator: genres
	// keep a black ring, artists get a light neutral ring, so the two kinds
	// stay distinguishable even within a single community's color group.
	const GENRE_BORDER_COLOR = "#000000";
	const ARTIST_BORDER_COLOR = "#e2e2f0";

	for (const v of genreVertices) {
		graph.addNode(genreNodeId(v.genre_id), {
			label: v.name,
			kind: "genre" satisfies NodeKind,
			genreId: v.genre_id,
			archiveCount: v.archive_count,
			borderColor: GENRE_BORDER_COLOR,
		});
	}

	for (const v of artistVertices) {
		graph.addNode(artistNodeId(v.artist_id), {
			label: v.name,
			kind: "artist" satisfies NodeKind,
			artistId: v.artist_id,
			archiveCount: v.archive_count,
			borderColor: ARTIST_BORDER_COLOR,
		});
	}

	const addWeightedEdge = (
		src: string,
		tgt: string,
		kind: "genre-genre" | "genre-artist" | "artist-artist",
		weight: number,
		maxWeight: number,
	): void => {
		if (!graph.hasNode(src) || !graph.hasNode(tgt) || graph.hasEdge(src, tgt))
			return;
		const opacity = edgeOpacity(weight, maxWeight);
		graph.addEdge(src, tgt, {
			kind,
			weight,
			size: edgeWidth(weight, maxWeight),
			color: `rgba(96, 96, 160, ${opacity.toFixed(2)})`,
		});
	};

	for (const e of genreGenreEdges) {
		addWeightedEdge(
			genreNodeId(e.source_genre_id),
			genreNodeId(e.target_genre_id),
			"genre-genre",
			e.shared_archive_count,
			maxGenreGenreWeight,
		);
	}

	for (const e of genreArtistEdges) {
		addWeightedEdge(
			genreNodeId(e.genre_id),
			artistNodeId(e.artist_id),
			"genre-artist",
			e.archive_count,
			maxGenreArtistWeight,
		);
	}

	for (const e of artistArtistEdges) {
		addWeightedEdge(
			artistNodeId(e.source_artist_id),
			artistNodeId(e.target_artist_id),
			"artist-artist",
			e.shared_song_count,
			maxArtistArtistWeight,
		);
	}

	// Detect communities over the full mixed graph, weighted by raw edge weight.
	louvain.assign(graph, { getEdgeWeight: "weight" });

	// Assign palette colors by community ID
	const communityIds = new Set<number>();
	graph.forEachNode((_, attrs) => communityIds.add(attrs.community as number));
	const colorMap = assignCommunityColors(communityIds, COMMUNITY_PALETTE);

	graph.forEachNode((node, attrs) => {
		const kind = attrs.kind as NodeKind;
		const archiveCount = attrs.archiveCount as number;
		const maxForKind = kind === "genre" ? maxGenreArchive : maxArtistArchive;
		const size = scaledNodeSize(archiveCount, maxForKind);

		graph.setNodeAttribute(node, "size", size);
		graph.setNodeAttribute(
			node,
			"color",
			colorMap.get(attrs.community as number) ?? "#888888",
		);
	});

	return graph;
}
