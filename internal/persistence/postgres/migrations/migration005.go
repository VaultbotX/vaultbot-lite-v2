package migrations

var Migration005 = &Migration{
	Name: "005-DuplicateSongLookup",
	Up: `
		CREATE TABLE duplicate_song_lookup (
		    source_song_spotify_id VARCHAR(255) NOT NULL,
		    target_song_spotify_id VARCHAR(255) NOT NULL
		);

		CREATE UNIQUE INDEX ON duplicate_song_lookup (source_song_spotify_id, target_song_spotify_id);

		INSERT INTO duplicate_song_lookup (source_song_spotify_id, target_song_spotify_id)
		SELECT spotify_id, spotify_id
		FROM songs;
	`,
	Down: `
 		DROP TABLE duplicate_song_lookup;
	`,
}
