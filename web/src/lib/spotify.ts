export function spotifyUrl(
	type: "artist" | "track" | "playlist",
	id: string,
): string {
	return `spotify:${type}:${id}`;
}
