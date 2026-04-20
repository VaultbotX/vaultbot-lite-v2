/**
 * Runs promises in parallel (like Promise.all) but accepts a named object
 * instead of an array, so results are keyed by name rather than by position.
 * This prevents destructuring order bugs when adding or reordering queries.
 *
 * @example
 * const { users, posts } = await allNamed({
 *   users: fetchUsers(),
 *   posts: fetchPosts(),
 * });
 */
export async function allNamed<T extends Record<string, Promise<unknown>>>(
	queries: T,
): Promise<{ [K in keyof T]: Awaited<T[K]> }> {
	const keys = Object.keys(queries) as (keyof T & string)[];
	const values = await Promise.all(keys.map((k) => queries[k]));
	return Object.fromEntries(keys.map((k, i) => [k, values[i]])) as {
		[K in keyof T]: Awaited<T[K]>;
	};
}

/**
 * Asserts the type of a database query result at the system boundary.
 * The neon template tag returns Record<string,any>[] which cannot be cast
 * directly to a concrete interface in TypeScript 6 — this helper confines
 * the necessary `as unknown as` cast to a single, documented location.
 *
 * @example
 * typed<User[]>(sql`SELECT id, name FROM users`)
 */
export function typed<T>(promise: Promise<unknown>): Promise<T> {
	return promise as unknown as Promise<T>;
}
