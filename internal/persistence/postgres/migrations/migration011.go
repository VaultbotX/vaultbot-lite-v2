package migrations

var Migration011 = &Migration{
	Name: "011-VSongsView",
	Up: `
		CREATE VIEW v_songs AS
		SELECT s.*
		FROM songs s
		JOIN duplicate_song_lookup dsl
		    ON s.spotify_id = dsl.source_song_spotify_id
		    AND s.spotify_id = dsl.target_song_spotify_id;
	`,
	Down: `
		DROP VIEW IF EXISTS v_songs;
	`,
}
