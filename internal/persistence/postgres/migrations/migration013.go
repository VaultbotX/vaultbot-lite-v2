package migrations

var Migration013 = &Migration{
	Name: "013-GraphVertexEdgeTimeRanges",
	Up: `
		DROP MATERIALIZED VIEW IF EXISTS artist_graph_edges;
		DROP MATERIALIZED VIEW IF EXISTS artist_graph_vertices;
		DROP MATERIALIZED VIEW IF EXISTS genre_graph_edges;
		DROP MATERIALIZED VIEW IF EXISTS genre_graph_vertices;

		CREATE MATERIALIZED VIEW genre_graph_vertices AS
		WITH canonical_events AS (
			SELECT
				dsl.target_song_spotify_id AS canonical_spotify_id,
				sa.created_at
			FROM song_archive sa
			JOIN songs raw ON sa.song_id = raw.id
			JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
		),
		artist_events AS (
			SELECT
				lsa.artist_id,
				ce.created_at
			FROM link_song_artists lsa
			JOIN v_songs vs ON vs.id = lsa.song_id
			JOIN canonical_events ce ON ce.canonical_spotify_id = vs.spotify_id
		),
		genre_events AS (
			SELECT
				lag.genre_id,
				ae.created_at
			FROM link_artist_genres lag
			JOIN artist_events ae ON ae.artist_id = lag.artist_id
		),
		genre_flagged AS (
			SELECT
				genre_id,
				created_at,
				(created_at - LAG(created_at) OVER (PARTITION BY genre_id ORDER BY created_at) > INTERVAL '24 hours'
					OR LAG(created_at) OVER (PARTITION BY genre_id ORDER BY created_at) IS NULL) AS is_new_island
			FROM genre_events
		),
		genre_islands AS (
			SELECT
				genre_id,
				created_at,
				SUM(is_new_island::int) OVER (PARTITION BY genre_id ORDER BY created_at) AS island_id
			FROM genre_flagged
		),
		genre_ranges AS (
			SELECT
				genre_id,
				jsonb_agg(jsonb_build_array(extract(epoch FROM min_ts)::bigint, extract(epoch FROM max_ts)::bigint) ORDER BY min_ts) AS ranges
			FROM (
				SELECT genre_id, island_id, MIN(created_at) AS min_ts, MAX(created_at) AS max_ts
				FROM genre_islands
				GROUP BY genre_id, island_id
			) collapsed
			GROUP BY genre_id
		),
		genre_counts AS (
			SELECT genre_id, COUNT(created_at)::int AS archive_count
			FROM genre_events
			GROUP BY genre_id
			HAVING COUNT(created_at) > 0
		)
		SELECT
			g.id AS genre_id,
			g.name,
			gc.archive_count,
			gr.ranges
		FROM genres g
		JOIN genre_counts gc ON gc.genre_id = g.id
		JOIN genre_ranges gr ON gr.genre_id = g.id;

		CREATE UNIQUE INDEX genre_graph_vertices_genre_id_idx ON genre_graph_vertices (genre_id);

		CREATE MATERIALIZED VIEW genre_graph_edges AS
		WITH canonical_events AS (
			SELECT
				dsl.target_song_spotify_id AS canonical_spotify_id,
				sa.created_at
			FROM song_archive sa
			JOIN songs raw ON sa.song_id = raw.id
			JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
		),
		artist_events AS (
			SELECT
				lsa.artist_id,
				ce.created_at
			FROM link_song_artists lsa
			JOIN v_songs vs ON vs.id = lsa.song_id
			JOIN canonical_events ce ON ce.canonical_spotify_id = vs.spotify_id
		),
		genre_pair_events AS (
			SELECT
				lag1.genre_id AS source_genre_id,
				lag2.genre_id AS target_genre_id,
				ae.created_at
			FROM link_artist_genres lag1
			JOIN link_artist_genres lag2
				ON lag2.artist_id = lag1.artist_id
				AND lag2.genre_id > lag1.genre_id
			JOIN artist_events ae ON ae.artist_id = lag1.artist_id
		),
		genre_pair_flagged AS (
			SELECT
				source_genre_id, target_genre_id, created_at,
				(created_at - LAG(created_at) OVER (PARTITION BY source_genre_id, target_genre_id ORDER BY created_at) > INTERVAL '24 hours'
					OR LAG(created_at) OVER (PARTITION BY source_genre_id, target_genre_id ORDER BY created_at) IS NULL) AS is_new_island
			FROM genre_pair_events
		),
		genre_pair_islands AS (
			SELECT
				source_genre_id, target_genre_id, created_at,
				SUM(is_new_island::int) OVER (PARTITION BY source_genre_id, target_genre_id ORDER BY created_at) AS island_id
			FROM genre_pair_flagged
		),
		genre_pair_ranges AS (
			SELECT
				source_genre_id, target_genre_id,
				jsonb_agg(jsonb_build_array(extract(epoch FROM min_ts)::bigint, extract(epoch FROM max_ts)::bigint) ORDER BY min_ts) AS ranges
			FROM (
				SELECT source_genre_id, target_genre_id, island_id, MIN(created_at) AS min_ts, MAX(created_at) AS max_ts
				FROM genre_pair_islands
				GROUP BY source_genre_id, target_genre_id, island_id
			) collapsed
			GROUP BY source_genre_id, target_genre_id
		),
		genre_pair_counts AS (
			SELECT source_genre_id, target_genre_id, COUNT(created_at)::int AS shared_archive_count
			FROM genre_pair_events
			GROUP BY source_genre_id, target_genre_id
			HAVING COUNT(created_at) > 0
		)
		SELECT
			gpc.source_genre_id,
			gpc.target_genre_id,
			gpc.shared_archive_count,
			gpr.ranges
		FROM genre_pair_counts gpc
		JOIN genre_pair_ranges gpr
			ON gpr.source_genre_id = gpc.source_genre_id AND gpr.target_genre_id = gpc.target_genre_id;

		CREATE UNIQUE INDEX genre_graph_edges_source_target_idx ON genre_graph_edges (source_genre_id, target_genre_id);
		CREATE INDEX genre_graph_edges_source_idx ON genre_graph_edges (source_genre_id);
		CREATE INDEX genre_graph_edges_target_idx ON genre_graph_edges (target_genre_id);

		CREATE MATERIALIZED VIEW artist_graph_vertices AS
		WITH canonical_events AS (
			SELECT
				dsl.target_song_spotify_id AS canonical_spotify_id,
				sa.created_at
			FROM song_archive sa
			JOIN songs raw ON sa.song_id = raw.id
			JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
		),
		artist_events AS (
			SELECT
				lsa.artist_id,
				ce.created_at
			FROM link_song_artists lsa
			JOIN v_songs vs ON vs.id = lsa.song_id
			JOIN canonical_events ce ON ce.canonical_spotify_id = vs.spotify_id
		),
		artist_flagged AS (
			SELECT
				artist_id, created_at,
				(created_at - LAG(created_at) OVER (PARTITION BY artist_id ORDER BY created_at) > INTERVAL '24 hours'
					OR LAG(created_at) OVER (PARTITION BY artist_id ORDER BY created_at) IS NULL) AS is_new_island
			FROM artist_events
		),
		artist_islands AS (
			SELECT
				artist_id, created_at,
				SUM(is_new_island::int) OVER (PARTITION BY artist_id ORDER BY created_at) AS island_id
			FROM artist_flagged
		),
		artist_ranges AS (
			SELECT
				artist_id,
				jsonb_agg(jsonb_build_array(extract(epoch FROM min_ts)::bigint, extract(epoch FROM max_ts)::bigint) ORDER BY min_ts) AS ranges
			FROM (
				SELECT artist_id, island_id, MIN(created_at) AS min_ts, MAX(created_at) AS max_ts
				FROM artist_islands
				GROUP BY artist_id, island_id
			) collapsed
			GROUP BY artist_id
		),
		artist_counts AS (
			SELECT artist_id, COUNT(created_at)::int AS archive_count
			FROM artist_events
			GROUP BY artist_id
			HAVING COUNT(created_at) > 0
		)
		SELECT
			a.id AS artist_id,
			a.name,
			ac.archive_count,
			ar.ranges
		FROM artists a
		JOIN artist_counts ac ON ac.artist_id = a.id
		JOIN artist_ranges ar ON ar.artist_id = a.id;

		CREATE UNIQUE INDEX artist_graph_vertices_artist_id_idx ON artist_graph_vertices (artist_id);

		CREATE MATERIALIZED VIEW artist_graph_edges AS
		WITH canonical_events AS (
			SELECT
				dsl.target_song_spotify_id AS canonical_spotify_id,
				sa.created_at
			FROM song_archive sa
			JOIN songs raw ON sa.song_id = raw.id
			JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
		),
		artist_pair_song_events AS (
			SELECT
				lsa1.artist_id AS source_artist_id,
				lsa2.artist_id AS target_artist_id,
				vs.id AS song_id,
				ce.created_at
			FROM link_song_artists lsa1
			JOIN link_song_artists lsa2
				ON lsa2.song_id = lsa1.song_id
				AND lsa2.artist_id > lsa1.artist_id
			JOIN v_songs vs ON vs.id = lsa1.song_id
			JOIN canonical_events ce ON ce.canonical_spotify_id = vs.spotify_id
		),
		artist_pair_flagged AS (
			SELECT
				source_artist_id, target_artist_id, created_at,
				(created_at - LAG(created_at) OVER (PARTITION BY source_artist_id, target_artist_id ORDER BY created_at) > INTERVAL '24 hours'
					OR LAG(created_at) OVER (PARTITION BY source_artist_id, target_artist_id ORDER BY created_at) IS NULL) AS is_new_island
			FROM artist_pair_song_events
		),
		artist_pair_islands AS (
			SELECT
				source_artist_id, target_artist_id, created_at,
				SUM(is_new_island::int) OVER (PARTITION BY source_artist_id, target_artist_id ORDER BY created_at) AS island_id
			FROM artist_pair_flagged
		),
		artist_pair_ranges AS (
			SELECT
				source_artist_id, target_artist_id,
				jsonb_agg(jsonb_build_array(extract(epoch FROM min_ts)::bigint, extract(epoch FROM max_ts)::bigint) ORDER BY min_ts) AS ranges
			FROM (
				SELECT source_artist_id, target_artist_id, island_id, MIN(created_at) AS min_ts, MAX(created_at) AS max_ts
				FROM artist_pair_islands
				GROUP BY source_artist_id, target_artist_id, island_id
			) collapsed
			GROUP BY source_artist_id, target_artist_id
		),
		artist_pair_counts AS (
			SELECT source_artist_id, target_artist_id, COUNT(DISTINCT song_id)::int AS shared_song_count
			FROM artist_pair_song_events
			GROUP BY source_artist_id, target_artist_id
			HAVING COUNT(DISTINCT song_id) > 0
		)
		SELECT
			apc.source_artist_id,
			apc.target_artist_id,
			apc.shared_song_count,
			apr.ranges
		FROM artist_pair_counts apc
		JOIN artist_pair_ranges apr
			ON apr.source_artist_id = apc.source_artist_id AND apr.target_artist_id = apc.target_artist_id;

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
