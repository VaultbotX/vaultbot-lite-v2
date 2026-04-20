import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

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

export const GET: RequestHandler = async ({ params }) => {
	// TODO(PR3): query artists and top tracks for genre params.id
	const data: GenreDetail = { genre_name: '', artists: [], tracks: [] };

	return json(data);
};
