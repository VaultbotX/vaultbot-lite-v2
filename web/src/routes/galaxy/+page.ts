import { parseNodeParam } from "$lib/graph";
import type { ArtistDetail } from "../api/artists/[id]/+server";
import type { GenreDetail } from "../api/genres/[id]/+server";
import type { GraphData } from "../api/graph/+server";
import type { PageLoad } from "./$types";

type InitialDetail =
	| { kind: "genre"; data: GenreDetail }
	| { kind: "artist"; data: ArtistDetail }
	| null;

export const load: PageLoad = async ({ fetch, url }) => {
	const node = parseNodeParam(url.searchParams.get("node"));

	const graphPromise = fetch("/api/graph").then((res) => {
		if (!res.ok) throw new Error(`Failed to load graph: ${res.status}`);
		return res.json() as Promise<GraphData>;
	});

	const detailPromise: Promise<InitialDetail> = node
		? fetch(
				node.kind === "genre"
					? `/api/genres/${node.id}`
					: `/api/artists/${node.id}`,
			).then((res) => {
				if (!res.ok) return null;
				return res
					.json()
					.then((data) => ({ kind: node.kind, data }) as InitialDetail);
			})
		: Promise.resolve(null);

	const [graphData, initialDetail] = await Promise.all([
		graphPromise,
		detailPromise,
	]);

	return {
		...graphData,
		initialNode: initialDetail ? node : null,
		initialDetail,
	};
};
