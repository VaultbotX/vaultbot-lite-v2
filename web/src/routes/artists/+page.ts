import type { ArtistsData } from "../api/artists/+server";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ fetch }) => {
	const res = await fetch("/api/artists");
	if (!res.ok) {
		throw new Error(`Failed to load artists: ${res.status}`);
	}
	const data: ArtistsData = await res.json();
	return data;
};
