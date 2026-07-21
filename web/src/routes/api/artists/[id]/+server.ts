import { neon } from "@neondatabase/serverless";
import { json } from "@sveltejs/kit";
import { allNamed, typed } from "$lib/allNamed";
import { parseTimeRangeParams, sqlBounds } from "$lib/graph";
import type { RequestHandler } from "./$types";

export interface ArtistSong {
	name: string;
	spotify_id: string;
	artist_ids: number[];
	artist_names: string[];
	artist_spotify_ids: string[];
	occurrences: number;
}

export interface ArtistGenre {
	genre_id: number;
	name: string;
}

export interface ConnectedArtist {
	artist_id: number;
	name: string;
	shared_song_count: number;
}

export interface ArtistDetail {
	artist_name: string;
	spotify_id: string;
	songs: ArtistSong[];
	genres: ArtistGenre[];
	connected_artists: ConnectedArtist[];
	rank: number;
	rank_total: number;
}

export const GET: RequestHandler = async ({ platform, params, url }) => {
	const artistId = Number(params.id);
	if (!Number.isInteger(artistId) || artistId <= 0) {
		return new Response("Invalid artist ID", { status: 400 });
	}

	const dbUrl = platform?.env?.DATABASE_URL;
	if (!dbUrl) {
		return new Response("DATABASE_URL not configured", { status: 500 });
	}

	const sql = neon(dbUrl);
	const range = parseTimeRangeParams(url.searchParams);
	const bounds = range ? sqlBounds(range) : null;

	const { artistRows, rankRows, songs, genres, connected_artists } = await allNamed({
		artistRows: typed<{ name: string; spotify_id: string }[]>(sql`
			SELECT name, spotify_id FROM artists WHERE id = ${artistId}
		`),
		rankRows: typed<{ rank: number; total: number }[]>(sql`
			SELECT rank, total FROM artist_rank WHERE artist_id = ${artistId}
		`),
		songs: bounds
			? typed<ArtistSong[]>(sql`
				WITH canonical_occurrences AS (
					SELECT dsl.target_song_spotify_id AS canonical_spotify_id, COUNT(sa.id)::int AS occurrences
					FROM song_archive sa
					JOIN songs raw ON raw.id = sa.song_id
					JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
					WHERE sa.created_at >= to_timestamp(${bounds.start})
						AND sa.created_at < to_timestamp(${bounds.endExclusive})
					GROUP BY dsl.target_song_spotify_id
				),
				song_all_artists AS (
					SELECT DISTINCT lsa.song_id, a.id AS artist_id, a.name, a.spotify_id
					FROM link_song_artists lsa
					JOIN artists a ON a.id = lsa.artist_id
				)
				SELECT
					s.name,
					s.spotify_id,
					array_agg(saa.artist_id ORDER BY saa.name) AS artist_ids,
					array_agg(saa.name ORDER BY saa.name) AS artist_names,
					array_agg(saa.spotify_id ORDER BY saa.name) AS artist_spotify_ids,
					co.occurrences
				FROM v_songs s
				JOIN link_song_artists lsa ON lsa.song_id = s.id AND lsa.artist_id = ${artistId}
				JOIN canonical_occurrences co ON co.canonical_spotify_id = s.spotify_id
				JOIN song_all_artists saa ON saa.song_id = s.id
				GROUP BY s.id, s.name, s.spotify_id, co.occurrences
				ORDER BY occurrences DESC
			`)
			: typed<ArtistSong[]>(sql`
				WITH canonical_occurrences AS (
					SELECT dsl.target_song_spotify_id AS canonical_spotify_id, COUNT(sa.id)::int AS occurrences
					FROM song_archive sa
					JOIN songs raw ON raw.id = sa.song_id
					JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
					GROUP BY dsl.target_song_spotify_id
				),
				song_all_artists AS (
					SELECT DISTINCT lsa.song_id, a.id AS artist_id, a.name, a.spotify_id
					FROM link_song_artists lsa
					JOIN artists a ON a.id = lsa.artist_id
				)
				SELECT
					s.name,
					s.spotify_id,
					array_agg(saa.artist_id ORDER BY saa.name) AS artist_ids,
					array_agg(saa.name ORDER BY saa.name) AS artist_names,
					array_agg(saa.spotify_id ORDER BY saa.name) AS artist_spotify_ids,
					co.occurrences
				FROM v_songs s
				JOIN link_song_artists lsa ON lsa.song_id = s.id AND lsa.artist_id = ${artistId}
				JOIN canonical_occurrences co ON co.canonical_spotify_id = s.spotify_id
				JOIN song_all_artists saa ON saa.song_id = s.id
				GROUP BY s.id, s.name, s.spotify_id, co.occurrences
				ORDER BY occurrences DESC
			`),
		genres: bounds
			? typed<ArtistGenre[]>(sql`
				SELECT DISTINCT g.id AS genre_id, g.name
				FROM genres g
				JOIN link_artist_genres lag ON lag.genre_id = g.id
				JOIN link_song_artists lsa ON lsa.artist_id = lag.artist_id
				JOIN song_archive sa ON sa.song_id = lsa.song_id
				WHERE lag.artist_id = ${artistId}
					AND sa.created_at >= to_timestamp(${bounds.start})
					AND sa.created_at < to_timestamp(${bounds.endExclusive})
				ORDER BY g.name
			`)
			: typed<ArtistGenre[]>(sql`
				SELECT g.id AS genre_id, g.name
				FROM genres g
				JOIN link_artist_genres lag ON lag.genre_id = g.id
				WHERE lag.artist_id = ${artistId}
				ORDER BY g.name
			`),
		connected_artists: bounds
			? typed<ConnectedArtist[]>(sql`
				WITH canonical_events AS (
					SELECT dsl.target_song_spotify_id AS canonical_spotify_id, sa.created_at
					FROM song_archive sa
					JOIN songs raw ON raw.id = sa.song_id
					JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
					WHERE sa.created_at >= to_timestamp(${bounds.start})
						AND sa.created_at < to_timestamp(${bounds.endExclusive})
				),
				artist_pair_song_events AS (
					SELECT
						lsa1.artist_id AS source_artist_id,
						lsa2.artist_id AS target_artist_id,
						vs.id AS song_id
					FROM link_song_artists lsa1
					JOIN link_song_artists lsa2
						ON lsa2.song_id = lsa1.song_id AND lsa2.artist_id != lsa1.artist_id
					JOIN v_songs vs ON vs.id = lsa1.song_id
					JOIN canonical_events ce ON ce.canonical_spotify_id = vs.spotify_id
					WHERE lsa1.artist_id = ${artistId}
				)
				SELECT a.id AS artist_id, a.name, COUNT(DISTINCT song_id)::int AS shared_song_count
				FROM artist_pair_song_events ape
				JOIN artists a ON a.id = ape.target_artist_id
				GROUP BY a.id, a.name
				ORDER BY shared_song_count DESC
				LIMIT 20
			`)
			: typed<ConnectedArtist[]>(sql`
				SELECT a.id AS artist_id, a.name, e.shared_song_count
				FROM artist_graph_edges e
				JOIN artists a ON a.id = e.target_artist_id
				WHERE e.source_artist_id = ${artistId}
				UNION ALL
				SELECT a.id AS artist_id, a.name, e.shared_song_count
				FROM artist_graph_edges e
				JOIN artists a ON a.id = e.source_artist_id
				WHERE e.target_artist_id = ${artistId}
				ORDER BY shared_song_count DESC
				LIMIT 20
			`),
	});

	if (artistRows.length === 0) {
		return new Response("Artist not found", { status: 404 });
	}

	return json({
		artist_name: artistRows[0].name,
		spotify_id: artistRows[0].spotify_id,
		songs,
		genres,
		connected_artists,
		rank: rankRows[0].rank,
		rank_total: rankRows[0].total,
	} satisfies ArtistDetail);
};
