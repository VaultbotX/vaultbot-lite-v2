// See https://svelte.dev/docs/kit/types#app.d.ts

declare global {
	namespace App {
		interface PageState {
			node?: { kind: "genre" | "artist"; id: number } | null;
		}
		interface Platform {
			env: {
				DATABASE_URL: string;
				SPOTIFY_PLAYLIST_ID: string;
				GENRE_SPOTIFY_PLAYLIST_ID: string;
				HIGH_SCORES_SPOTIFY_PLAYLIST_ID: string;
				THROWBACK_SPOTIFY_PLAYLIST_ID: string;
				VARIETY_SPOTIFY_PLAYLIST_ID: string;
			};
			context: {
				waitUntil(promise: Promise<unknown>): void;
			};
			caches: CacheStorage & { default: Cache };
		}
	}
}

export {};
