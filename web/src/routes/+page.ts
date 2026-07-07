import { allNamed } from "$lib/allNamed";
import type { PageLoad } from "./$types";
import type { PlaylistIds } from "./api/playlists/+server";
import type { StatsData } from "./api/stats/+server";

export const load: PageLoad = async ({ fetch }) => {
	const { stats, playlists } = await allNamed({
		stats: (async () => {
			const res = await fetch("/api/stats");
			if (!res.ok) throw new Error(`Failed to load stats: ${res.status}`);
			return (await res.json()) as StatsData;
		})(),
		playlists: (async () => {
			const res = await fetch("/api/playlists");
			if (!res.ok) throw new Error(`Failed to load playlists: ${res.status}`);
			return (await res.json()) as PlaylistIds;
		})(),
	});

	return { ...stats, playlists };
};
