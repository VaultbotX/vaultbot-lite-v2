import { describe, expect, it } from "vitest";
import { spotifyUrl } from "./spotify";

describe("spotifyUrl", () => {
	it("builds a deeplink for an artist", () => {
		expect(spotifyUrl("artist", "123")).toBe("spotify:artist:123");
	});

	it("builds a deeplink for a track", () => {
		expect(spotifyUrl("track", "456")).toBe("spotify:track:456");
	});

	it("builds a deeplink for a playlist", () => {
		expect(spotifyUrl("playlist", "789")).toBe("spotify:playlist:789");
	});
});
