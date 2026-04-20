import { neon } from "@neondatabase/serverless";
import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export interface GenreArtist {
	name: string;
	spotify_id: string;
	archive_count: number;
}

export interface GenreTrack {
	name: string;
	spotify_id: string;
	artist_names: string[];
	artist_spotify_ids: string[];
	occurrences: number;
}

export interface GenreDetail {
	genre_name: string;
	artists: GenreArtist[];
	tracks: GenreTrack[];
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

	const [genreRows, artists, tracks] = await Promise.all([
		sql`
			SELECT name FROM genres WHERE id = ${genreId}
		`,
		sql`
			SELECT a.name, a.spotify_id, COUNT(sa.id)::int AS archive_count
			FROM artists a
			JOIN link_artist_genres lag ON lag.artist_id = a.id
			JOIN link_song_artists lsa ON lsa.artist_id = a.id
			JOIN song_archive sa ON sa.song_id = lsa.song_id
			WHERE lag.genre_id = ${genreId}
			GROUP BY a.id, a.name, a.spotify_id
			ORDER BY archive_count DESC
			LIMIT 20
		`,
		sql`
			WITH song_occurrences AS (
				SELECT song_id, COUNT(id)::int AS occurrences
				FROM song_archive
				GROUP BY song_id
			),
			song_artists AS (
				SELECT DISTINCT lsa.song_id, a.name, a.spotify_id
				FROM link_song_artists lsa
				JOIN artists a ON a.id = lsa.artist_id
			)
			SELECT
				s.name,
				s.spotify_id,
				array_agg(sa.name ORDER BY sa.name) AS artist_names,
				array_agg(sa.spotify_id ORDER BY sa.name) AS artist_spotify_ids,
				so.occurrences
			FROM songs s
			JOIN link_song_genres lsg ON lsg.song_id = s.id
			JOIN song_occurrences so ON so.song_id = s.id
			JOIN song_artists sa ON sa.song_id = s.id
			WHERE lsg.genre_id = ${genreId}
			GROUP BY s.id, s.name, s.spotify_id, so.occurrences
			ORDER BY occurrences DESC
			LIMIT 20
		`,
	]);

	if (genreRows.length === 0) {
		return new Response("Genre not found", { status: 404 });
	}

	return json({
		genre_name: (genreRows[0] as { name: string }).name,
		artists: artists as unknown as GenreArtist[],
		tracks: tracks as unknown as GenreTrack[],
	} satisfies GenreDetail);
};
