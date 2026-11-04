import { env } from "cloudflare:workers";
import { neon } from "@neondatabase/serverless";
import { json } from "@sveltejs/kit";
import { typed } from "#lib/allNamed.ts";
import {
	parseRankingsParam,
	parseTimeRangeParams,
	type RankingItem,
	sqlBounds,
} from "#lib/graph.ts";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({ url }) => {
	const kind = parseRankingsParam(url.searchParams.get("kind"));
	if (!kind) {
		return new Response(
			"Invalid or missing 'kind' parameter (must be 'genres' or 'artists')",
			{ status: 400 },
		);
	}

	const dbUrl = env.DATABASE_URL;
	if (!dbUrl) {
		return new Response("DATABASE_URL not configured", { status: 500 });
	}

	const sql = neon(dbUrl);
	const range = parseTimeRangeParams(url.searchParams);
	const bounds = range ? sqlBounds(range) : null;

	if (kind === "genres") {
		const items = bounds
			? await typed<RankingItem[]>(sql`
				WITH canonical_events AS (
					SELECT dsl.target_song_spotify_id AS canonical_spotify_id
					FROM song_archive sa
					JOIN songs raw ON sa.song_id = raw.id
					JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
					WHERE sa.created_at >= to_timestamp(${bounds.start})
						AND sa.created_at < to_timestamp(${bounds.endExclusive})
				),
				artist_events AS (
					SELECT lsa.artist_id
					FROM link_song_artists lsa
					JOIN v_songs vs ON vs.id = lsa.song_id
					JOIN canonical_events ce ON ce.canonical_spotify_id = vs.spotify_id
				),
				genre_events AS (
					SELECT lag.genre_id
					FROM link_artist_genres lag
					JOIN artist_events ae ON ae.artist_id = lag.artist_id
				)
				SELECT
					g.id,
					g.name,
					gr.rank,
					COUNT(ge.genre_id)::int AS entries
				FROM genre_events ge
				JOIN genres g ON g.id = ge.genre_id
				JOIN genre_rank gr ON gr.genre_id = g.id
				GROUP BY g.id, g.name, gr.rank
				ORDER BY gr.rank ASC
			`)
			: await typed<RankingItem[]>(sql`
				SELECT
					gv.genre_id AS id,
					gv.name,
					gr.rank,
					gv.archive_count AS entries
				FROM genre_graph_vertices gv
				JOIN genre_rank gr ON gr.genre_id = gv.genre_id
				ORDER BY gr.rank ASC
			`);

		return json(items, {
			headers: bounds
				? { "Cache-Control": "public, max-age=300" }
				: { "Cache-Control": "public, max-age=21600" },
		});
	}

	const items = bounds
		? await typed<RankingItem[]>(sql`
			WITH canonical_events AS (
				SELECT dsl.target_song_spotify_id AS canonical_spotify_id
				FROM song_archive sa
				JOIN songs raw ON sa.song_id = raw.id
				JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
				WHERE sa.created_at >= to_timestamp(${bounds.start})
					AND sa.created_at < to_timestamp(${bounds.endExclusive})
			)
			SELECT
				a.id,
				a.name,
				ar.rank,
				COUNT(ce.canonical_spotify_id)::int AS entries
			FROM canonical_events ce
			JOIN v_songs vs ON vs.spotify_id = ce.canonical_spotify_id
			JOIN link_song_artists lsa ON lsa.song_id = vs.id
			JOIN artists a ON a.id = lsa.artist_id
			JOIN artist_rank ar ON ar.artist_id = a.id
			GROUP BY a.id, a.name, ar.rank
			ORDER BY ar.rank ASC
		`)
		: await typed<RankingItem[]>(sql`
			SELECT
				av.artist_id AS id,
				av.name,
				ar.rank,
				av.archive_count AS entries
			FROM artist_graph_vertices av
			JOIN artist_rank ar ON ar.artist_id = av.artist_id
			ORDER BY ar.rank ASC
		`);

	return json(items, {
		headers: bounds
			? { "Cache-Control": "public, max-age=300" }
			: { "Cache-Control": "public, max-age=21600" },
	});
};
