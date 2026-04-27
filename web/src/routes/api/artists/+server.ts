import { neon } from "@neondatabase/serverless";
import { json } from "@sveltejs/kit";
import { typed } from "$lib/allNamed";
import type { RequestHandler } from "./$types";

export interface ArtistSummary {
	artist_id: number;
	name: string;
	spotify_id: string;
	unique_songs: number;
	archive_count: number;
	genre_count: number;
}

export interface ArtistsData {
	artists: ArtistSummary[];
}

export const GET: RequestHandler = async ({ platform }) => {
	const dbUrl = platform?.env?.DATABASE_URL;
	if (!dbUrl) {
		return new Response("DATABASE_URL not configured", { status: 500 });
	}

	const sql = neon(dbUrl);

	const artists = await typed<ArtistSummary[]>(sql`
		WITH canonical_archive AS (
			SELECT
				dsl.target_song_spotify_id AS canonical_spotify_id,
				COUNT(sa.id)::int AS occ
			FROM song_archive sa
			JOIN songs raw ON sa.song_id = raw.id
			JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
			GROUP BY dsl.target_song_spotify_id
		)
		SELECT
			a.id AS artist_id,
			a.name,
			a.spotify_id,
			COUNT(DISTINCT vs.id)::int AS unique_songs,
			COALESCE(SUM(ca.occ), 0)::int AS archive_count,
			COUNT(DISTINCT lag.genre_id)::int AS genre_count
		FROM artists a
		JOIN link_song_artists lsa ON lsa.artist_id = a.id
		JOIN v_songs vs ON vs.id = lsa.song_id
		LEFT JOIN canonical_archive ca ON ca.canonical_spotify_id = vs.spotify_id
		LEFT JOIN link_artist_genres lag ON lag.artist_id = a.id
		GROUP BY a.id, a.name, a.spotify_id
		ORDER BY archive_count DESC
	`);

	return json({ artists } satisfies ArtistsData, {
		headers: { "Cache-Control": "public, max-age=3600, s-maxage=3600" },
	});
};
