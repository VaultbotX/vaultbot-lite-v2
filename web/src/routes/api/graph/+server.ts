import { neon } from "@neondatabase/serverless";
import { json } from "@sveltejs/kit";
import { allNamed, typed } from "$lib/allNamed";
import type { RequestHandler } from "./$types";

export interface GenreVertex {
	genre_id: number;
	name: string;
	artist_count: number;
}

export interface GenreEdge {
	source_genre_id: number;
	target_genre_id: number;
	shared_artist_count: number;
}

export interface GraphData {
	vertices: GenreVertex[];
	edges: GenreEdge[];
}

export const GET: RequestHandler = async ({ platform, url }) => {
	const dbUrl = platform?.env?.DATABASE_URL;
	if (!dbUrl) {
		return new Response("DATABASE_URL not configured", { status: 500 });
	}

	const sql = neon(dbUrl);
	const dynamic = url.searchParams.get("dynamic") === "true";

	const { vertices, edges } = dynamic
		? await allNamed({
				vertices: typed<GenreVertex[]>(sql`
					SELECT g.id AS genre_id, g.name, COUNT(DISTINCT lag.artist_id) AS artist_count
					FROM genres g
					JOIN link_artist_genres lag ON lag.genre_id = g.id
					WHERE lag.artist_id IN (
						SELECT DISTINCT lsa.artist_id
						FROM link_song_artists lsa
						JOIN song_archive sa ON sa.song_id = lsa.song_id
						WHERE sa.created_at > NOW() - INTERVAL '14 days'
					)
					GROUP BY g.id, g.name
					HAVING COUNT(DISTINCT lag.artist_id) > 0
					ORDER BY artist_count DESC
				`),
				edges: typed<GenreEdge[]>(sql`
					SELECT
						lag1.genre_id AS source_genre_id,
						lag2.genre_id AS target_genre_id,
						COUNT(DISTINCT lag1.artist_id) AS shared_artist_count
					FROM link_artist_genres lag1
					JOIN link_artist_genres lag2
						ON lag2.artist_id = lag1.artist_id
						AND lag2.genre_id > lag1.genre_id
					WHERE lag1.artist_id IN (
						SELECT DISTINCT lsa.artist_id
						FROM link_song_artists lsa
						JOIN song_archive sa ON sa.song_id = lsa.song_id
						WHERE sa.created_at > NOW() - INTERVAL '14 days'
					)
					GROUP BY lag1.genre_id, lag2.genre_id
					HAVING COUNT(DISTINCT lag1.artist_id) > 0
				`),
			})
		: await allNamed({
				vertices: typed<GenreVertex[]>(sql`
					SELECT genre_id, name, artist_count
					FROM genre_graph_vertices
					ORDER BY artist_count DESC
				`),
				edges: typed<GenreEdge[]>(sql`
					SELECT source_genre_id, target_genre_id, shared_artist_count
					FROM genre_graph_edges
				`),
			});

	return json({ vertices, edges } satisfies GraphData, {
		headers: { "Cache-Control": `public, max-age=${dynamic ? 300 : 21600}` },
	});
};
