package migrations

var Migration014 = &Migration{
	Name: "014-ArtistGenreRankMaterializedViews",
	Up: `
		CREATE MATERIALIZED VIEW artist_rank AS
		SELECT
			artist_id,
			RANK() OVER (ORDER BY archive_count DESC)::int AS rank,
			COUNT(*) OVER ()::int AS total
		FROM artist_graph_vertices;

		CREATE UNIQUE INDEX artist_rank_artist_id_idx ON artist_rank (artist_id);

		CREATE MATERIALIZED VIEW genre_rank AS
		SELECT
			genre_id,
			RANK() OVER (ORDER BY archive_count DESC)::int AS rank,
			COUNT(*) OVER ()::int AS total
		FROM genre_graph_vertices;

		CREATE UNIQUE INDEX genre_rank_genre_id_idx ON genre_rank (genre_id);
	`,
	Down: `
		DROP MATERIALIZED VIEW IF EXISTS genre_rank;
		DROP MATERIALIZED VIEW IF EXISTS artist_rank;
	`,
}
