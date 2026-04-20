import type { GenreDetail } from "../../api/genre/[id]/+server";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ fetch, params }) => {
	const res = await fetch(`/api/genre/${params.id}`);
	if (res.status === 404) {
		return {
			notFound: true,
			genre_name: "",
			artists: [],
			tracks: [],
			connected_genres: [],
		};
	}
	if (!res.ok) {
		throw new Error(`Failed to load genre: ${res.status}`);
	}
	const detail: GenreDetail = await res.json();
	return { ...detail, notFound: false };
};
