import type { GraphData } from "../api/graph/+server";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ fetch }) => {
	const res = await fetch("/api/graph");
	if (!res.ok) throw new Error(`Failed to load graph: ${res.status}`);
	const data: GraphData = await res.json();
	return data;
};
