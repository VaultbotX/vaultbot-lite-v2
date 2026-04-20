import { neon } from "@neondatabase/serverless";
import { json } from "@sveltejs/kit";
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

export const GET: RequestHandler = async ({ platform }) => {
	const dbUrl = platform?.env?.DATABASE_URL;
	if (!dbUrl) {
		return new Response("DATABASE_URL not configured", { status: 500 });
	}

	const sql = neon(dbUrl);

	const [vertices, edges] = await Promise.all([
		sql<GenreVertex[]>`
			SELECT genre_id, name, artist_count
			FROM genre_graph_vertices
			ORDER BY artist_count DESC
		`,
		sql<GenreEdge[]>`
			SELECT source_genre_id, target_genre_id, shared_artist_count
			FROM genre_graph_edges
		`,
	]);

	return json({ vertices, edges } satisfies GraphData, {
		headers: { "Cache-Control": "public, max-age=21600" },
	});
};
