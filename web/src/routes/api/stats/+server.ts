import { neon } from "@neondatabase/serverless";
import { json } from "@sveltejs/kit";
import { allNamed, typed } from "$lib/allNamed";
import type { RequestHandler } from "./$types";

export interface MonthlyCount {
	month: string;
	count: number;
}

export interface ArtistCount {
	name: string;
	song_count: number;
}

export interface GenreCount {
	genre_id: number;
	name: string;
	song_count: number;
}

interface SummaryRow {
	total_songs: number;
	total_archive_entries: number;
	total_artists: number;
	total_genres: number;
}

export interface StatsData {
	generated_at: string;
	summary: {
		total_songs: number;
		total_archive_entries: number;
		total_artists: number;
		total_genres: number;
	};
	songs_over_time: MonthlyCount[];
	top_artists: ArtistCount[];
	genre_distribution: GenreCount[];
}

export const GET: RequestHandler = async ({ platform }) => {
	const dbUrl = platform?.env?.DATABASE_URL;
	if (!dbUrl) {
		return new Response("DATABASE_URL not configured", { status: 500 });
	}

	const sql = neon(dbUrl);

	const { summaryRows, songs_over_time, top_artists, genre_distribution } =
		await allNamed({
			summaryRows: typed<SummaryRow[]>(sql`
				SELECT
					(SELECT COUNT(*)::int FROM songs)        AS total_songs,
					(SELECT COUNT(*)::int FROM song_archive) AS total_archive_entries,
					(SELECT COUNT(*)::int FROM artists)      AS total_artists,
					(SELECT COUNT(*)::int FROM genres)       AS total_genres
			`),
			songs_over_time: typed<MonthlyCount[]>(sql`
				SELECT
					TO_CHAR(DATE_TRUNC('month', created_at), 'YYYY-MM') AS month,
					COUNT(*)::int                                        AS count
				FROM song_archive
				GROUP BY DATE_TRUNC('month', created_at)
				ORDER BY DATE_TRUNC('month', created_at)
			`),
			top_artists: typed<ArtistCount[]>(sql`
				SELECT a.name, COUNT(DISTINCT lsa.song_id)::int AS song_count
				FROM artists a
				JOIN link_song_artists lsa ON a.id = lsa.artist_id
				GROUP BY a.id, a.name
				ORDER BY song_count DESC
				LIMIT 15
			`),
			genre_distribution: typed<GenreCount[]>(sql`
				SELECT g.id AS genre_id, g.name, COUNT(DISTINCT lsg.song_id)::int AS song_count
				FROM genres g
				JOIN link_song_genres lsg ON g.id = lsg.genre_id
				GROUP BY g.id, g.name
				ORDER BY song_count DESC
				LIMIT 30
			`),
		});

	return json({
		generated_at: new Date().toISOString(),
		summary: summaryRows[0],
		songs_over_time,
		top_artists,
		genre_distribution,
	} satisfies StatsData);
};
