import { describe, expect, it, vi } from "vitest";
import type { RequestEvent } from "./$types";

vi.mock("cloudflare:workers", () => ({
	env: { DATABASE_URL: "postgres://mock" },
}));

import { GET } from "./+server";

function createMockEvent(url: string): RequestEvent {
	return {
		url: new URL(url),
	} as unknown as RequestEvent;
}

describe("GET /api/rankings", () => {
	it("returns 400 when kind is missing", async () => {
		const response = await GET(
			createMockEvent("http://localhost/api/rankings"),
		);
		expect(response.status).toBe(400);
		expect(await response.text()).toContain(
			"Invalid or missing 'kind' parameter",
		);
	});

	it("returns 400 when kind is invalid", async () => {
		const response = await GET(
			createMockEvent("http://localhost/api/rankings?kind=albums"),
		);
		expect(response.status).toBe(400);
	});
});
