import { neon } from "@neondatabase/serverless";
import { json } from "@sveltejs/kit";
import { allNamed, typed } from "$lib/allNamed";
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

export interface ArtistDetail {
	artist_name: string;
	spotify_id: string;
	songs: ArtistSong[];
	genres: ArtistGenre[];
}

export const GET: RequestHandler = async ({ platform, params }) => {
	const artistId = Number(params.id);
	if (!Number.isInteger(artistId) || artistId <= 0) {
		return new Response("Invalid artist ID", { status: 400 });
	}

	const dbUrl = platform?.env?.DATABASE_URL;
	if (!dbUrl) {
		return new Response("DATABASE_URL not configured", { status: 500 });
	}

	const sql = neon(dbUrl);

	const { artistRows, songs, genres } = await allNamed({
		artistRows: typed<{ name: string; spotify_id: string }[]>(sql`
			SELECT name, spotify_id FROM artists WHERE id = ${artistId}
		`),
		songs: typed<ArtistSong[]>(sql`
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
		genres: typed<ArtistGenre[]>(sql`
			SELECT g.id AS genre_id, g.name
			FROM genres g
			JOIN link_artist_genres lag ON lag.genre_id = g.id
			WHERE lag.artist_id = ${artistId}
			ORDER BY g.name
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
	} satisfies ArtistDetail);
};
