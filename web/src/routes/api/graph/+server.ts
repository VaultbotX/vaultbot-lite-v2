import { neon } from "@neondatabase/serverless";
import { json } from "@sveltejs/kit";
import { allNamed, typed } from "$lib/allNamed";
import type { TimeRange } from "$lib/graph";
import type { RequestHandler } from "./$types";

export interface GenreVertex {
	genre_id: number;
	name: string;
	archive_count: number;
	ranges: TimeRange[];
}

export interface ArtistVertex {
	artist_id: number;
	name: string;
	archive_count: number;
	ranges: TimeRange[];
}

export interface GenreGenreEdge {
	source_genre_id: number;
	target_genre_id: number;
	shared_archive_count: number;
	ranges: TimeRange[];
}

export interface GenreArtistEdge {
	genre_id: number;
	artist_id: number;
	archive_count: number;
	ranges: TimeRange[];
}

export interface ArtistArtistEdge {
	source_artist_id: number;
	target_artist_id: number;
	shared_song_count: number;
	ranges: TimeRange[];
}

export interface GraphData {
	genreVertices: GenreVertex[];
	artistVertices: ArtistVertex[];
	genreGenreEdges: GenreGenreEdge[];
	genreArtistEdges: GenreArtistEdge[];
	artistArtistEdges: ArtistArtistEdge[];
}

export const GET: RequestHandler = async ({ platform }) => {
	const dbUrl = platform?.env?.DATABASE_URL;
	if (!dbUrl) {
		return new Response("DATABASE_URL not configured", { status: 500 });
	}

	const sql = neon(dbUrl);

	const data = await allNamed({
		genreVertices: typed<GenreVertex[]>(sql`
			SELECT genre_id, name, archive_count, ranges
			FROM genre_graph_vertices
			ORDER BY archive_count DESC
		`),
		artistVertices: typed<ArtistVertex[]>(sql`
			SELECT artist_id, name, archive_count, ranges
			FROM artist_graph_vertices
			ORDER BY archive_count DESC
		`),
		genreGenreEdges: typed<GenreGenreEdge[]>(sql`
			SELECT source_genre_id, target_genre_id, shared_archive_count, ranges
			FROM genre_graph_edges
		`),
		genreArtistEdges: typed<GenreArtistEdge[]>(sql`
			SELECT lag.genre_id, lag.artist_id, agv.archive_count, agv.ranges
			FROM link_artist_genres lag
			JOIN artist_graph_vertices agv ON agv.artist_id = lag.artist_id
		`),
		artistArtistEdges: typed<ArtistArtistEdge[]>(sql`
			SELECT source_artist_id, target_artist_id, shared_song_count, ranges
			FROM artist_graph_edges
		`),
	});

	return json(data satisfies GraphData, {
		headers: { "Cache-Control": "public, max-age=21600" },
	});
};
