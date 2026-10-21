package songs

import (
	"database/sql"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/zmb3/spotify/v2"
)

type Song struct {
	Id             int       `db:"id"`
	SpotifyId      string    `db:"spotify_id"`
	Name           string    `db:"name"`
	ReleaseDate    time.Time `db:"release_date"`
	SpotifyAlbumId string    `db:"spotify_album_id"`
	CreatedAt      time.Time `db:"created_at"`
	Duration       int       `db:"duration"`
	Popularity     int       `db:"popularity"`
	AlbumName      string    `db:"album_name"`
}

// AddSong adds a song to the database
func AddSong(tx *sqlx.Tx, track *spotify.FullTrack, genreIds []int, artistIds []int) (Song, error) {
	var addSong Song

	err := tx.QueryRowx(`
		SELECT id,
		       spotify_id,
		       name,
		       release_date,
		       spotify_album_id,
		       created_at,
		       duration,
		       popularity,
		       album_name
		FROM songs
		WHERE spotify_id = $1
	`, track.ID.String()).StructScan(&addSong)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return Song{}, err
		}

		err := tx.QueryRowx(`
			INSERT INTO songs (
			                   spotify_id,
			                   name,
			                   release_date,
			                   spotify_album_id,
			                   duration,
			                   popularity,
			                   album_name
		    )
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (spotify_id) DO NOTHING
			RETURNING id, created_at
		`, track.ID.String(),
			track.Name,
			track.Album.ReleaseDateTime(),
			track.Album.ID.String(),
			int(track.Duration),
			track.Popularity,
			track.Album.Name).StructScan(&addSong)

		if err != nil {
			return Song{}, err
		}

		addSong.SpotifyId = track.ID.String()
		addSong.Name = track.Name
		addSong.ReleaseDate = track.Album.ReleaseDateTime()
		addSong.SpotifyAlbumId = track.Album.ID.String()

		_, err = tx.Exec(`
			INSERT INTO duplicate_song_lookup (source_song_spotify_id, target_song_spotify_id)
			VALUES ($1, $1)
			ON CONFLICT DO NOTHING
		`, track.ID.String())
		if err != nil {
			return Song{}, err
		}
	}

	// insert link records
	for _, genreId := range genreIds {
		_, err := tx.Exec(`
			INSERT INTO link_song_genres (song_id, genre_id) 
			VALUES ($1, $2)
			ON CONFLICT (song_id, genre_id) DO NOTHING
		`, addSong.Id, genreId)
		if err != nil {
			return Song{}, err
		}
	}

	for _, artistId := range artistIds {
		_, err := tx.Exec(`
			INSERT INTO link_song_artists (song_id, artist_id) 
			VALUES ($1, $2)
			ON CONFLICT (song_id, artist_id) DO NOTHING
		`, addSong.Id, artistId)
		if err != nil {
			return Song{}, err
		}
	}

	return addSong, nil
}

// GetOverallTopSongs retrieves the top songs based on their occurrence in the song_archive table
func GetOverallTopSongs(db *sqlx.DB, limit int) ([]Song, error) {
	var songs []Song

	err := db.Select(&songs, `
		SELECT s.id,
			   s.spotify_id,
			   s.name,
			   s.release_date,
			   s.spotify_album_id,
			   s.created_at,
			   s.duration,
			   s.popularity,
			   s.album_name
		FROM song_archive sa
				 JOIN songs raw ON sa.song_id = raw.id
				 JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
				 JOIN songs s ON s.spotify_id = dsl.target_song_spotify_id
		GROUP BY s.id
		ORDER BY COUNT(sa.id) DESC
		LIMIT $1;
	`, limit)

	if err != nil {
		return nil, err
	}

	return songs, nil
}

// GetTopSongsByWindow retrieves the top songs added to the archive during a randomly selected window of time
// of duration windowDuration, bounded between the earliest archived song and min(latest archived song, now - bufferDuration).
// If the selected window contains fewer than minCount tracks, it retries up to maxAttempts times.
// Returns the selected songs, the window start and end time, or nil and zero times if no valid window is found.
func GetTopSongsByWindow(db *sqlx.DB, windowDuration, bufferDuration time.Duration, minCount, limit, maxAttempts int) ([]Song, time.Time, time.Time, error) {
	var bounds struct {
		MinCreatedAt sql.NullTime `db:"min_created_at"`
		MaxCreatedAt sql.NullTime `db:"max_created_at"`
	}

	err := db.Get(&bounds, `
		SELECT MIN(created_at) AS min_created_at,
		       MAX(created_at) AS max_created_at
		FROM song_archive
	`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, time.Time{}, time.Time{}, nil
		}
		return nil, time.Time{}, time.Time{}, err
	}

	if !bounds.MinCreatedAt.Valid || !bounds.MaxCreatedAt.Valid {
		return nil, time.Time{}, time.Time{}, nil
	}

	minDate := bounds.MinCreatedAt.Time
	maxDate := bounds.MaxCreatedAt.Time
	now := time.Now()

	latestPossibleEnd := maxDate
	if bufferLimit := now.Add(-bufferDuration); bufferLimit.Before(latestPossibleEnd) {
		latestPossibleEnd = bufferLimit
	}

	latestPossibleStart := latestPossibleEnd.Add(-windowDuration)
	if latestPossibleStart.Before(minDate) {
		// Not enough history to form the required window and buffer
		return nil, time.Time{}, time.Time{}, nil
	}

	rangeSeconds := int64(latestPossibleStart.Sub(minDate).Seconds())

	for attempt := 0; attempt < maxAttempts; attempt++ {
		var windowStart time.Time
		if rangeSeconds <= 0 {
			windowStart = minDate
		} else {
			offset := rand.Int64N(rangeSeconds + 1)
			windowStart = minDate.Add(time.Duration(offset) * time.Second)
		}
		windowEnd := windowStart.Add(windowDuration)

		var songs []Song
		err = db.Select(&songs, `
			SELECT s.id,
			       s.spotify_id,
			       s.name,
			       s.release_date,
			       s.spotify_album_id,
			       s.created_at,
			       s.duration,
			       s.popularity,
			       s.album_name
			FROM song_archive sa
			         JOIN songs raw ON sa.song_id = raw.id
			         JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
			         JOIN songs s ON s.spotify_id = dsl.target_song_spotify_id
			WHERE sa.created_at >= $1 AND sa.created_at < $2
			GROUP BY s.id
			ORDER BY COUNT(sa.id) DESC
			LIMIT $3;
		`, windowStart, windowEnd, limit)
		if err != nil {
			return nil, time.Time{}, time.Time{}, err
		}

		if len(songs) >= minCount {
			return songs, windowStart, windowEnd, nil
		}
	}

	return nil, time.Time{}, time.Time{}, nil
}

// GetRandomSongs retrieves a random selection of songs from the archive.
func GetRandomSongs(db *sqlx.DB, limit int) ([]Song, error) {
	var songs []Song
	err := db.Select(&songs, `
		SELECT s.id,
		       s.spotify_id,
		       s.name,
		       s.release_date,
		       s.spotify_album_id,
		       s.created_at,
		       s.duration,
		       s.popularity,
		       s.album_name
		FROM song_archive sa
		         JOIN songs raw ON sa.song_id = raw.id
		         JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
		         JOIN songs s ON s.spotify_id = dsl.target_song_spotify_id
		GROUP BY s.id
		ORDER BY RANDOM()
		LIMIT $1;
	`, limit)
	if err != nil {
		return nil, err
	}
	return songs, nil
}

// GetTopSongsByGenre retrieves the top 100 songs for a given genre based on their occurrence in the song_archive table
func GetTopSongsByGenre(db *sqlx.DB, genreId int) ([]Song, error) {
	var songs []Song

	err := db.Select(&songs, `
		SELECT s.id,
			   s.spotify_id,
			   s.name,
			   s.release_date,
			   s.spotify_album_id,
			   s.created_at,
			   s.duration,
			   s.popularity,
			   s.album_name
		FROM song_archive sa
				 JOIN songs raw ON sa.song_id = raw.id
				 JOIN duplicate_song_lookup dsl ON dsl.source_song_spotify_id = raw.spotify_id
				 JOIN songs s ON s.spotify_id = dsl.target_song_spotify_id
				 JOIN link_song_genres lsg ON s.id = lsg.song_id
		WHERE lsg.genre_id = $1
		GROUP BY s.id
		ORDER BY COUNT(sa.id) DESC
		LIMIT 100;
	`, genreId)

	if err != nil {
		return nil, err
	}

	return songs, nil
}
