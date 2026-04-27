import type { ArtistDetail } from "../../api/artists/[id]/+server";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ fetch, params }) => {
	const res = await fetch(`/api/artists/${params.id}`);
	if (res.status === 404) {
		return {
			notFound: true,
			artist_name: "",
			spotify_id: "",
			songs: [],
			genres: [],
		};
	}
	if (!res.ok) {
		throw new Error(`Failed to load artist: ${res.status}`);
	}
	const detail: ArtistDetail = await res.json();
	return { ...detail, notFound: false };
};
