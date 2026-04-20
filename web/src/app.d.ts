// See https://svelte.dev/docs/kit/types#app.d.ts

declare global {
	namespace App {
		interface Platform {
			env: {
				DATABASE_URL: string;
			};
			context: {
				waitUntil(promise: Promise<unknown>): void;
			};
			caches: CacheStorage & { default: Cache };
		}
	}
}

export {};
