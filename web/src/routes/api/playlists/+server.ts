import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export interface PlaylistIds {
	dynamic: string;
	genre: string;
	highscores: string;
	throwback: string;
	variety: string;
}

export const GET: RequestHandler = async ({ platform }) => {
	const env = platform?.env;
	if (
		!env?.SPOTIFY_PLAYLIST_ID ||
		!env?.GENRE_SPOTIFY_PLAYLIST_ID ||
		!env?.HIGH_SCORES_SPOTIFY_PLAYLIST_ID ||
		!env?.THROWBACK_SPOTIFY_PLAYLIST_ID ||
		!env?.VARIETY_SPOTIFY_PLAYLIST_ID
	) {
		return new Response("Playlist IDs not configured", { status: 500 });
	}

	return json({
		dynamic: env.SPOTIFY_PLAYLIST_ID,
		genre: env.GENRE_SPOTIFY_PLAYLIST_ID,
		highscores: env.HIGH_SCORES_SPOTIFY_PLAYLIST_ID,
		throwback: env.THROWBACK_SPOTIFY_PLAYLIST_ID,
		variety: env.VARIETY_SPOTIFY_PLAYLIST_ID,
	} satisfies PlaylistIds);
};
