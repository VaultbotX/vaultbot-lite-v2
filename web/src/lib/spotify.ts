export function spotifyUrl(type: "artist" | "track", id: string): string {
	return `spotify:${type}:${id}`;
}
