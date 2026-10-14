// Bindings are read via `import { env } from "cloudflare:workers"`: adapter-cloudflare 8
// no longer populates `event.platform.env`. Declared locally (not via a global
// `@cloudflare/workers-types` reference) because those globals retype `Response.json()`.
declare module "cloudflare:workers" {
	export const env: {
		DATABASE_URL: string;
		SPOTIFY_PLAYLIST_ID: string;
		GENRE_SPOTIFY_PLAYLIST_ID: string;
		HIGH_SCORES_SPOTIFY_PLAYLIST_ID: string;
		THROWBACK_SPOTIFY_PLAYLIST_ID: string;
		VARIETY_SPOTIFY_PLAYLIST_ID: string;
	};
}
