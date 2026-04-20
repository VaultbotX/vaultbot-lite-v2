import type { PageLoad } from "./$types";
import type { GraphData } from "./api/graph/+server";

export const load: PageLoad = async ({ fetch }) => {
	const res = await fetch("/api/graph");
	const data: GraphData = await res.json();
	return data;
};
