import { neon } from "@neondatabase/serverless";
import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export interface GenreArtist {
	name: string;
	archive_count: number;
}

export interface GenreTrack {
	name: string;
	artist_names: string[];
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
			SELECT a.name, COUNT(sa.id)::int AS archive_count
			FROM artists a
			JOIN link_artist_genres lag ON lag.artist_id = a.id
			JOIN link_song_artists lsa ON lsa.artist_id = a.id
			JOIN song_archive sa ON sa.song_id = lsa.song_id
			WHERE lag.genre_id = ${genreId}
			GROUP BY a.id, a.name
			ORDER BY archive_count DESC
			LIMIT 20
		`,
		sql`
			SELECT
				s.name,
				array_agg(DISTINCT a.name ORDER BY a.name) AS artist_names,
				COUNT(sa.id)::int AS occurrences
			FROM songs s
			JOIN link_song_genres lsg ON lsg.song_id = s.id
			JOIN song_archive sa ON sa.song_id = s.id
			JOIN link_song_artists lsa ON lsa.song_id = s.id
			JOIN artists a ON a.id = lsa.artist_id
			WHERE lsg.genre_id = ${genreId}
			GROUP BY s.id, s.name
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
