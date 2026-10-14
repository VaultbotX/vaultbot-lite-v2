// See https://svelte.dev/docs/kit/types#app.d.ts

declare global {
	namespace App {
		interface PageState {
			node?: { kind: "genre" | "artist"; id: number } | null;
		}
		interface Platform {
			context: {
				waitUntil(promise: Promise<unknown>): void;
			};
			caches: CacheStorage & { default: Cache };
		}
	}
}

export {};
