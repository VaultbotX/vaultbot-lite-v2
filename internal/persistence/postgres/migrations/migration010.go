package migrations

var Migration010 = &Migration{
	Name: "010-GenreGraphMaterializedViews",
	Up: `
		CREATE MATERIALIZED VIEW genre_graph_vertices AS
		SELECT g.id AS genre_id, g.name, COUNT(DISTINCT lag.artist_id) AS artist_count
		FROM genres g
		JOIN link_artist_genres lag ON lag.genre_id = g.id
		GROUP BY g.id, g.name
		HAVING COUNT(DISTINCT lag.artist_id) > 0;

		CREATE UNIQUE INDEX genre_graph_vertices_genre_id_idx ON genre_graph_vertices (genre_id);

		CREATE MATERIALIZED VIEW genre_graph_edges AS
		SELECT
			lag1.genre_id AS source_genre_id,
			lag2.genre_id AS target_genre_id,
			COUNT(DISTINCT lag1.artist_id) AS shared_artist_count
		FROM link_artist_genres lag1
		JOIN link_artist_genres lag2
			ON lag2.artist_id = lag1.artist_id
			AND lag2.genre_id > lag1.genre_id
		GROUP BY lag1.genre_id, lag2.genre_id
		HAVING COUNT(DISTINCT lag1.artist_id) > 0;

		CREATE UNIQUE INDEX genre_graph_edges_source_target_idx ON genre_graph_edges (source_genre_id, target_genre_id);
		CREATE INDEX genre_graph_edges_source_idx ON genre_graph_edges (source_genre_id);
		CREATE INDEX genre_graph_edges_target_idx ON genre_graph_edges (target_genre_id);
	`,
	Down: `
		DROP MATERIALIZED VIEW IF EXISTS genre_graph_edges;
		DROP MATERIALIZED VIEW IF EXISTS genre_graph_vertices;
	`,
}
