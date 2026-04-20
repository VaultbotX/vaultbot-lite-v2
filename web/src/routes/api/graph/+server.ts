import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

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

export const GET: RequestHandler = async () => {
	// TODO(PR3): query genre_graph_vertices and genre_graph_edges materialized views
	const data: GraphData = { vertices: [], edges: [] };

	return json(data, {
		headers: { 'Cache-Control': 'public, max-age=21600' }
	});
};
