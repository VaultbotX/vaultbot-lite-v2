package songs

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/zmb3/spotify/v2"
	"time"
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
