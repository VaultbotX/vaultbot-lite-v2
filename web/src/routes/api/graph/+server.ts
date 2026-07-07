import { neon } from "@neondatabase/serverless";
import { json } from "@sveltejs/kit";
import { allNamed, typed } from "$lib/allNamed";
import type { RequestHandler } from "./$types";

export interface GenreVertex {
	genre_id: number;
	name: string;
	archive_count: number;
}

export interface ArtistVertex {
	artist_id: number;
	name: string;
	archive_count: number;
}

export interface GenreGenreEdge {
	source_genre_id: number;
	target_genre_id: number;
	shared_archive_count: number;
}

export interface GenreArtistEdge {
	genre_id: number;
	artist_id: number;
	archive_count: number;
}

export interface ArtistArtistEdge {
	source_artist_id: number;
	target_artist_id: number;
	shared_song_count: number;
}

export interface GraphData {
	genreVertices: GenreVertex[];
	artistVertices: ArtistVertex[];
	genreGenreEdges: GenreGenreEdge[];
	genreArtistEdges: GenreArtistEdge[];
	artistArtistEdges: ArtistArtistEdge[];
}

const CANONICAL_ARCHIVE_CTE = `
	canonical_archive AS (
		SELECT
			dsl.target_song_spotify_id AS canonical_spotify_id,
			COUNT(sa.id)::int AS occ
		FROM song_archive sa
		JOIN songs raw ON sa.song_id = raw.id
		JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
		GROUP BY dsl.target_song_spotify_id
	)
`;

export const GET: RequestHandler = async ({ platform, url }) => {
	const dbUrl = platform?.env?.DATABASE_URL;
	if (!dbUrl) {
		return new Response("DATABASE_URL not configured", { status: 500 });
	}

	const sql = neon(dbUrl);
	const dynamic = url.searchParams.get("dynamic") === "true";

	const data = dynamic
		? await allNamed({
				genreVertices: typed<GenreVertex[]>(sql`
					WITH ${sql.unsafe(CANONICAL_ARCHIVE_CTE)},
					recent_artist_ids AS (
						SELECT DISTINCT lsa.artist_id
						FROM song_archive sa
						JOIN link_song_artists lsa ON lsa.song_id = sa.song_id
						WHERE sa.created_at >= NOW() - INTERVAL '14 days'
					),
					artist_archive AS (
						SELECT
							lsa.artist_id,
							COALESCE(SUM(ca.occ), 0)::int AS archive_count
						FROM link_song_artists lsa
						JOIN v_songs vs ON vs.id = lsa.song_id
						LEFT JOIN canonical_archive ca ON ca.canonical_spotify_id = vs.spotify_id
						WHERE lsa.artist_id IN (SELECT artist_id FROM recent_artist_ids)
						GROUP BY lsa.artist_id
					)
					SELECT g.id AS genre_id, g.name, COALESCE(SUM(aa.archive_count), 0)::int AS archive_count
					FROM genres g
					JOIN link_artist_genres lag ON lag.genre_id = g.id
					JOIN artist_archive aa ON aa.artist_id = lag.artist_id
					GROUP BY g.id, g.name
					HAVING COALESCE(SUM(aa.archive_count), 0) > 0
					ORDER BY archive_count DESC
				`),
				artistVertices: typed<ArtistVertex[]>(sql`
					WITH ${sql.unsafe(CANONICAL_ARCHIVE_CTE)},
					recent_artist_ids AS (
						SELECT DISTINCT lsa.artist_id
						FROM song_archive sa
						JOIN link_song_artists lsa ON lsa.song_id = sa.song_id
						WHERE sa.created_at >= NOW() - INTERVAL '14 days'
					),
					artist_archive AS (
						SELECT
							lsa.artist_id,
							COALESCE(SUM(ca.occ), 0)::int AS archive_count
						FROM link_song_artists lsa
						JOIN v_songs vs ON vs.id = lsa.song_id
						LEFT JOIN canonical_archive ca ON ca.canonical_spotify_id = vs.spotify_id
						WHERE lsa.artist_id IN (SELECT artist_id FROM recent_artist_ids)
						GROUP BY lsa.artist_id
					)
					SELECT a.id AS artist_id, a.name, aa.archive_count
					FROM artists a
					JOIN artist_archive aa ON aa.artist_id = a.id
					WHERE aa.archive_count > 0
					ORDER BY aa.archive_count DESC
				`),
				genreGenreEdges: typed<GenreGenreEdge[]>(sql`
					WITH ${sql.unsafe(CANONICAL_ARCHIVE_CTE)},
					recent_artist_ids AS (
						SELECT DISTINCT lsa.artist_id
						FROM song_archive sa
						JOIN link_song_artists lsa ON lsa.song_id = sa.song_id
						WHERE sa.created_at >= NOW() - INTERVAL '14 days'
					),
					artist_archive AS (
						SELECT
							lsa.artist_id,
							COALESCE(SUM(ca.occ), 0)::int AS archive_count
						FROM link_song_artists lsa
						JOIN v_songs vs ON vs.id = lsa.song_id
						LEFT JOIN canonical_archive ca ON ca.canonical_spotify_id = vs.spotify_id
						WHERE lsa.artist_id IN (SELECT artist_id FROM recent_artist_ids)
						GROUP BY lsa.artist_id
					)
					SELECT
						lag1.genre_id AS source_genre_id,
						lag2.genre_id AS target_genre_id,
						COALESCE(SUM(aa.archive_count), 0)::int AS shared_archive_count
					FROM link_artist_genres lag1
					JOIN link_artist_genres lag2
						ON lag2.artist_id = lag1.artist_id
						AND lag2.genre_id > lag1.genre_id
					JOIN artist_archive aa ON aa.artist_id = lag1.artist_id
					GROUP BY lag1.genre_id, lag2.genre_id
					HAVING COALESCE(SUM(aa.archive_count), 0) > 0
				`),
				genreArtistEdges: typed<GenreArtistEdge[]>(sql`
					WITH ${sql.unsafe(CANONICAL_ARCHIVE_CTE)},
					recent_artist_ids AS (
						SELECT DISTINCT lsa.artist_id
						FROM song_archive sa
						JOIN link_song_artists lsa ON lsa.song_id = sa.song_id
						WHERE sa.created_at >= NOW() - INTERVAL '14 days'
					),
					artist_archive AS (
						SELECT
							lsa.artist_id,
							COALESCE(SUM(ca.occ), 0)::int AS archive_count
						FROM link_song_artists lsa
						JOIN v_songs vs ON vs.id = lsa.song_id
						LEFT JOIN canonical_archive ca ON ca.canonical_spotify_id = vs.spotify_id
						WHERE lsa.artist_id IN (SELECT artist_id FROM recent_artist_ids)
						GROUP BY lsa.artist_id
					)
					SELECT lag.genre_id, lag.artist_id, aa.archive_count
					FROM link_artist_genres lag
					JOIN artist_archive aa ON aa.artist_id = lag.artist_id
					WHERE aa.archive_count > 0
				`),
				artistArtistEdges: typed<ArtistArtistEdge[]>(sql`
					WITH recent_artist_ids AS (
						SELECT DISTINCT lsa.artist_id
						FROM song_archive sa
						JOIN link_song_artists lsa ON lsa.song_id = sa.song_id
						WHERE sa.created_at >= NOW() - INTERVAL '14 days'
					)
					SELECT
						lsa1.artist_id AS source_artist_id,
						lsa2.artist_id AS target_artist_id,
						COUNT(DISTINCT vs.id)::int AS shared_song_count
					FROM link_song_artists lsa1
					JOIN link_song_artists lsa2
						ON lsa2.song_id = lsa1.song_id
						AND lsa2.artist_id > lsa1.artist_id
					JOIN v_songs vs ON vs.id = lsa1.song_id
					WHERE lsa1.artist_id IN (SELECT artist_id FROM recent_artist_ids)
					  AND lsa2.artist_id IN (SELECT artist_id FROM recent_artist_ids)
					GROUP BY lsa1.artist_id, lsa2.artist_id
					HAVING COUNT(DISTINCT vs.id) > 0
				`),
			})
		: await allNamed({
				genreVertices: typed<GenreVertex[]>(sql`
					SELECT genre_id, name, archive_count
					FROM genre_graph_vertices
					ORDER BY archive_count DESC
				`),
				artistVertices: typed<ArtistVertex[]>(sql`
					SELECT artist_id, name, archive_count
					FROM artist_graph_vertices
					ORDER BY archive_count DESC
				`),
				genreGenreEdges: typed<GenreGenreEdge[]>(sql`
					SELECT source_genre_id, target_genre_id, shared_archive_count
					FROM genre_graph_edges
				`),
				genreArtistEdges: typed<GenreArtistEdge[]>(sql`
					SELECT lag.genre_id, lag.artist_id, agv.archive_count
					FROM link_artist_genres lag
					JOIN artist_graph_vertices agv ON agv.artist_id = lag.artist_id
				`),
				artistArtistEdges: typed<ArtistArtistEdge[]>(sql`
					SELECT source_artist_id, target_artist_id, shared_song_count
					FROM artist_graph_edges
				`),
			});

	return json(data satisfies GraphData, {
		headers: { "Cache-Control": `public, max-age=${dynamic ? 300 : 21600}` },
	});
};
