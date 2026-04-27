import { neon } from "@neondatabase/serverless";
import { json } from "@sveltejs/kit";
import { allNamed, typed } from "$lib/allNamed";
import type { RequestHandler } from "./$types";

export interface GenreArtist {
	artist_id: number;
	name: string;
	spotify_id: string;
	archive_count: number;
}

export interface GenreTrack {
	name: string;
	spotify_id: string;
	artist_ids: number[];
	artist_names: string[];
	artist_spotify_ids: string[];
	occurrences: number;
}

export interface ConnectedGenre {
	genre_id: number;
	name: string;
	shared_artist_count: number;
}

export interface GenreDetail {
	genre_name: string;
	artists: GenreArtist[];
	tracks: GenreTrack[];
	connected_genres: ConnectedGenre[];
}

export const GET: RequestHandler = async ({ platform, params }) => {
	const genreId = Number(params.id);
	if (!Number.isInteger(genreId) || genreId <= 0) {
		return new Response("Invalid genre ID", { status: 400 });
	}

	const dbUrl = platform?.env?.DATABASE_URL;
	if (!dbUrl) {
		return new Response("DATABASE_URL not configured", { status: 500 });
	}

	const sql = neon(dbUrl);

	const { genreRows, artists, tracks, connected_genres } = await allNamed({
		genreRows: typed<{ name: string }[]>(sql`
			SELECT name FROM genres WHERE id = ${genreId}
		`),
		artists: typed<GenreArtist[]>(sql`
			SELECT a.id AS artist_id, a.name, a.spotify_id, COUNT(sa.id)::int AS archive_count
			FROM artists a
			JOIN link_artist_genres lag ON lag.artist_id = a.id
			JOIN link_song_artists lsa ON lsa.artist_id = a.id
			JOIN song_archive sa ON sa.song_id = lsa.song_id
			WHERE lag.genre_id = ${genreId}
			GROUP BY a.id, a.name, a.spotify_id
			ORDER BY archive_count DESC
			LIMIT 20
		`),
		tracks: typed<GenreTrack[]>(sql`
			WITH canonical_occurrences AS (
				SELECT dsl.target_song_spotify_id AS canonical_spotify_id, COUNT(sa.id)::int AS occurrences
				FROM song_archive sa
				JOIN songs raw ON raw.id = sa.song_id
				JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
				GROUP BY dsl.target_song_spotify_id
			),
			song_artists AS (
				SELECT DISTINCT lsa.song_id, a.id AS artist_id, a.name, a.spotify_id
				FROM link_song_artists lsa
				JOIN artists a ON a.id = lsa.artist_id
			)
			SELECT
				s.name,
				s.spotify_id,
				array_agg(sa.artist_id ORDER BY sa.name) AS artist_ids,
				array_agg(sa.name ORDER BY sa.name) AS artist_names,
				array_agg(sa.spotify_id ORDER BY sa.name) AS artist_spotify_ids,
				co.occurrences
			FROM v_songs s
			JOIN link_song_genres lsg ON lsg.song_id = s.id
			JOIN canonical_occurrences co ON co.canonical_spotify_id = s.spotify_id
			JOIN song_artists sa ON sa.song_id = s.id
			WHERE lsg.genre_id = ${genreId}
			GROUP BY s.id, s.name, s.spotify_id, co.occurrences
			ORDER BY occurrences DESC
			LIMIT 20
		`),
		connected_genres: typed<ConnectedGenre[]>(sql`
			SELECT g.id AS genre_id, g.name, e.shared_artist_count
			FROM genre_graph_edges e
			JOIN genres g ON g.id = e.target_genre_id
			WHERE e.source_genre_id = ${genreId}
			UNION ALL
			SELECT g.id AS genre_id, g.name, e.shared_artist_count
			FROM genre_graph_edges e
			JOIN genres g ON g.id = e.source_genre_id
			WHERE e.target_genre_id = ${genreId}
			ORDER BY shared_artist_count DESC
		`),
	});

	if (genreRows.length === 0) {
		return new Response("Genre not found", { status: 404 });
	}

	return json({
		genre_name: genreRows[0].name,
		artists,
		tracks,
		connected_genres,
	} satisfies GenreDetail);
};
