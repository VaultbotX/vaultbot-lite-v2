package migrations

var Migration012 = &Migration{
	Name: "012-MixedGenreArtistGraphMaterializedViews",
	Up: `
		DROP MATERIALIZED VIEW IF EXISTS genre_graph_edges;
		DROP MATERIALIZED VIEW IF EXISTS genre_graph_vertices;

		CREATE MATERIALIZED VIEW genre_graph_vertices AS
		WITH canonical_archive AS (
			SELECT
				dsl.target_song_spotify_id AS canonical_spotify_id,
				COUNT(sa.id)::int AS occ
			FROM song_archive sa
			JOIN songs raw ON sa.song_id = raw.id
			JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
			GROUP BY dsl.target_song_spotify_id
		),
		artist_archive AS (
			SELECT
				lsa.artist_id,
				COALESCE(SUM(ca.occ), 0)::int AS archive_count
			FROM link_song_artists lsa
			JOIN v_songs vs ON vs.id = lsa.song_id
			LEFT JOIN canonical_archive ca ON ca.canonical_spotify_id = vs.spotify_id
			GROUP BY lsa.artist_id
		)
		SELECT g.id AS genre_id, g.name, COALESCE(SUM(aa.archive_count), 0)::int AS archive_count
		FROM genres g
		JOIN link_artist_genres lag ON lag.genre_id = g.id
		JOIN artist_archive aa ON aa.artist_id = lag.artist_id
		GROUP BY g.id, g.name
		HAVING COALESCE(SUM(aa.archive_count), 0) > 0;

		CREATE UNIQUE INDEX genre_graph_vertices_genre_id_idx ON genre_graph_vertices (genre_id);

		CREATE MATERIALIZED VIEW genre_graph_edges AS
		WITH canonical_archive AS (
			SELECT
				dsl.target_song_spotify_id AS canonical_spotify_id,
				COUNT(sa.id)::int AS occ
			FROM song_archive sa
			JOIN songs raw ON sa.song_id = raw.id
			JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
			GROUP BY dsl.target_song_spotify_id
		),
		artist_archive AS (
			SELECT
				lsa.artist_id,
				COALESCE(SUM(ca.occ), 0)::int AS archive_count
			FROM link_song_artists lsa
			JOIN v_songs vs ON vs.id = lsa.song_id
			LEFT JOIN canonical_archive ca ON ca.canonical_spotify_id = vs.spotify_id
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
		HAVING COALESCE(SUM(aa.archive_count), 0) > 0;

		CREATE UNIQUE INDEX genre_graph_edges_source_target_idx ON genre_graph_edges (source_genre_id, target_genre_id);
		CREATE INDEX genre_graph_edges_source_idx ON genre_graph_edges (source_genre_id);
		CREATE INDEX genre_graph_edges_target_idx ON genre_graph_edges (target_genre_id);

		CREATE MATERIALIZED VIEW artist_graph_vertices AS
		WITH canonical_archive AS (
			SELECT
				dsl.target_song_spotify_id AS canonical_spotify_id,
				COUNT(sa.id)::int AS occ
			FROM song_archive sa
			JOIN songs raw ON sa.song_id = raw.id
			JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
			GROUP BY dsl.target_song_spotify_id
		)
		SELECT
			a.id AS artist_id,
			a.name,
			COALESCE(SUM(ca.occ), 0)::int AS archive_count
		FROM artists a
		JOIN link_song_artists lsa ON lsa.artist_id = a.id
		JOIN v_songs vs ON vs.id = lsa.song_id
		LEFT JOIN canonical_archive ca ON ca.canonical_spotify_id = vs.spotify_id
		GROUP BY a.id, a.name
		HAVING COALESCE(SUM(ca.occ), 0) > 0;

		CREATE UNIQUE INDEX artist_graph_vertices_artist_id_idx ON artist_graph_vertices (artist_id);

		CREATE MATERIALIZED VIEW artist_graph_edges AS
		SELECT
			lsa1.artist_id AS source_artist_id,
			lsa2.artist_id AS target_artist_id,
			COUNT(DISTINCT vs.id)::int AS shared_song_count
		FROM link_song_artists lsa1
		JOIN link_song_artists lsa2
			ON lsa2.song_id = lsa1.song_id
			AND lsa2.artist_id > lsa1.artist_id
		JOIN v_songs vs ON vs.id = lsa1.song_id
		GROUP BY lsa1.artist_id, lsa2.artist_id
		HAVING COUNT(DISTINCT vs.id) > 0;

		CREATE UNIQUE INDEX artist_graph_edges_source_target_idx ON artist_graph_edges (source_artist_id, target_artist_id);
		CREATE INDEX artist_graph_edges_source_idx ON artist_graph_edges (source_artist_id);
		CREATE INDEX artist_graph_edges_target_idx ON artist_graph_edges (target_artist_id);
	`,
	Down: `
		DROP MATERIALIZED VIEW IF EXISTS artist_graph_edges;
		DROP MATERIALIZED VIEW IF EXISTS artist_graph_vertices;
		DROP MATERIALIZED VIEW IF EXISTS genre_graph_edges;
		DROP MATERIALIZED VIEW IF EXISTS genre_graph_vertices;
	`,
}
