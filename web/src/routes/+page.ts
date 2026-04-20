import type { PageLoad } from "./$types";
import type { StatsData } from "./api/stats/+server";

export const load: PageLoad = async ({ fetch }) => {
	const res = await fetch("/api/stats");
	if (!res.ok) throw new Error(`Failed to load stats: ${res.status}`);
	const data: StatsData = await res.json();
	return data;
};
