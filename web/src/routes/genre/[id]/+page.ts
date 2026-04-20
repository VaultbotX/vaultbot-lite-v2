import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, params }) => {
	const res = await fetch(`/api/genre/${params.id}`);
	const data = await res.json();
	return { ...data, genreId: params.id };
};
