package migrations

var Migration_001 = &Migration{
	Name: "001-Initial",
	Up: `
		CREATE TABLE IF NOT EXISTS songs (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			spotify_id VARCHAR(64) NOT NULL,
			name VARCHAR(255) NOT NULL,
			release_date DATE,
			spotify_album_id VARCHAR(64) NOT NULL,
			acousticness FLOAT,
			danceability FLOAT,
			duration_ms INTEGER,
			energy FLOAT,
			instrumentalness FLOAT,
			key INTEGER,
			liveness FLOAT,
			loudness FLOAT,
			mode INTEGER,
			speechiness FLOAT,
			tempo FLOAT,
			time_signature INTEGER,
			valence FLOAT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE UNIQUE INDEX IF NOT EXISTS idx_songs_spotify_id ON songs(spotify_id);
		
		CREATE TABLE IF NOT EXISTS genres (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE UNIQUE INDEX IF NOT EXISTS idx_genres_name ON genres(name);
		
		CREATE TABLE IF NOT EXISTS link_song_genres (
			song_id INTEGER NOT NULL,
			genre_id INTEGER NOT NULL,
			PRIMARY KEY (song_id, genre_id),
			FOREIGN KEY (song_id) REFERENCES songs(id),
			FOREIGN KEY (genre_id) REFERENCES genres(id)
		);
		
		CREATE TABLE IF NOT EXISTS artists (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			spotify_id VARCHAR(64) NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE UNIQUE INDEX IF NOT EXISTS idx_artists_spotify_id ON artists(spotify_id);
		
		CREATE TABLE IF NOT EXISTS link_song_artists (
			song_id INTEGER NOT NULL,
			artist_id INTEGER NOT NULL,
			PRIMARY KEY (song_id, artist_id),
			FOREIGN KEY (song_id) REFERENCES songs(id),
			FOREIGN KEY (artist_id) REFERENCES artists(id)
		);
		
		CREATE TABLE IF NOT EXISTS link_artist_genres (
			artist_id INTEGER NOT NULL,
			genre_id INTEGER NOT NULL,
			PRIMARY KEY (artist_id, genre_id),
			FOREIGN KEY (artist_id) REFERENCES artists(id),
			FOREIGN KEY (genre_id) REFERENCES genres(id)
		);
		
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			discord_id VARCHAR(64) NOT NULL,
			discord_username VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_discord_id ON users(discord_id);
		
		CREATE TABLE IF NOT EXISTS song_archive (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			song_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (song_id) REFERENCES songs(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		
		CREATE INDEX IF NOT EXISTS idx_song_archive_created_at ON song_archive(created_at);
		`,
	Down: `
		DROP TABLE IF EXISTS song_archive;
		DROP TABLE IF EXISTS users;
		DROP TABLE IF EXISTS link_artist_genres;
		DROP TABLE IF EXISTS link_song_artists;
		DROP TABLE IF EXISTS artists;
		DROP TABLE IF EXISTS link_song_genres;
		DROP TABLE IF EXISTS genres;
		DROP TABLE IF EXISTS songs;
	`,
}
