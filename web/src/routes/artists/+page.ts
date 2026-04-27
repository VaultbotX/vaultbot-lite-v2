import type { ArtistsData } from "../api/artists/+server";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ fetch, url }) => {
	const page = Math.max(1, Number(url.searchParams.get("page") ?? "1"));
	const q = url.searchParams.get("q") ?? "";
	const current = url.searchParams.get("current") === "1";

	const params = new URLSearchParams({ page: String(page) });
	if (q) params.set("q", q);
	if (current) params.set("current", "1");

	const res = await fetch(`/api/artists?${params}`);
	if (!res.ok) {
		throw new Error(`Failed to load artists: ${res.status}`);
	}
	const data: ArtistsData = await res.json();
	return { ...data, q, current };
};
