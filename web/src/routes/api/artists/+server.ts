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
	total: number;
	page: number;
	pageSize: number;
}

export const GET: RequestHandler = async ({ platform, url }) => {
	const dbUrl = platform?.env?.DATABASE_URL;
	if (!dbUrl) {
		return new Response("DATABASE_URL not configured", { status: 500 });
	}

	const pageSize = 50;
	const page = Math.max(1, Number(url.searchParams.get("page") ?? "1"));
	const offset = (page - 1) * pageSize;
	const q = url.searchParams.get("q")?.trim() ?? "";
	const current = url.searchParams.get("current") === "1";
	const searchParam = q ? `%${q}%` : null;

	const sql = neon(dbUrl);

	type ArtistRow = ArtistSummary & { total_count: number };

	const rows = await typed<ArtistRow[]>(sql`
		WITH canonical_archive AS (
			SELECT
				dsl.target_song_spotify_id AS canonical_spotify_id,
				COUNT(sa.id)::int AS occ
			FROM song_archive sa
			JOIN songs raw ON sa.song_id = raw.id
			JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
			GROUP BY dsl.target_song_spotify_id
		),
		recent_artist_ids AS (
			SELECT DISTINCT lsa.artist_id
			FROM song_archive sa
			JOIN link_song_artists lsa ON lsa.song_id = sa.song_id
			WHERE sa.created_at >= NOW() - INTERVAL '14 days'
		),
		artist_songs AS (
			SELECT
				lsa.artist_id,
				COUNT(DISTINCT vs.id)::int AS unique_songs,
				COALESCE(SUM(ca.occ), 0)::int AS archive_count
			FROM link_song_artists lsa
			JOIN v_songs vs ON vs.id = lsa.song_id
			LEFT JOIN canonical_archive ca ON ca.canonical_spotify_id = vs.spotify_id
			GROUP BY lsa.artist_id
		),
		artist_genres AS (
			SELECT artist_id, COUNT(DISTINCT genre_id)::int AS genre_count
			FROM link_artist_genres
			GROUP BY artist_id
		),
		filtered AS (
			SELECT
				a.id AS artist_id,
				a.name,
				a.spotify_id,
				aso.unique_songs,
				aso.archive_count,
				COALESCE(ag.genre_count, 0) AS genre_count
			FROM artists a
			JOIN artist_songs aso ON aso.artist_id = a.id
			LEFT JOIN artist_genres ag ON ag.artist_id = a.id
			WHERE (NOT ${current} OR a.id IN (SELECT artist_id FROM recent_artist_ids))
			  AND (${searchParam}::text IS NULL OR a.name ILIKE ${searchParam})
		)
		SELECT *, COUNT(*) OVER()::int AS total_count
		FROM filtered
		ORDER BY archive_count DESC
		LIMIT ${pageSize} OFFSET ${offset}
	`);

	const total = rows[0]?.total_count ?? 0;
	const artists = rows.map(({ artist_id, name, spotify_id, unique_songs, archive_count, genre_count }) => ({
		artist_id,
		name,
		spotify_id,
		unique_songs,
		archive_count,
		genre_count,
	})) satisfies ArtistSummary[];

	return json({ artists, total, page, pageSize } satisfies ArtistsData, {
		headers: { "Cache-Control": "no-store" },
	});
};
