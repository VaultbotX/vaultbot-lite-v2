package songs

import (
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
}

// AddSong adds a song to the database
func AddSong(tx *sqlx.Tx, track *spotify.FullTrack, genreIds []int, artistIds []int) (Song, error) {
	row, err := tx.NamedExec(`
		INSERT INTO songs (spotify_id, name, release_date, spotify_album_id) 
		VALUES (:spotify_id, :name, :release_date, :spotify_album_id)
		ON CONFLICT (spotify_id) DO NOTHING
	`, map[string]any{
		"spotify_id":       track.ID.String(),
		"name":             track.Name,
		"release_date":     track.Album.ReleaseDate,
		"spotify_album_id": track.Album.ID.String(),
	})

	if err != nil {
		return Song{}, err
	}

	id, err := row.LastInsertId()
	if err != nil {
		return Song{}, err
	}

	// insert link records
	for _, genreId := range genreIds {
		_, err := tx.Exec(`
			INSERT INTO link_song_genres (song_id, genre_id) 
			VALUES ($1, $2)
			ON CONFLICT (song_id, genre_id) DO NOTHING
		`, id, genreId)
		if err != nil {
			return Song{}, err
		}
	}

	for _, artistId := range artistIds {
		_, err := tx.Exec(`
			INSERT INTO link_song_artists (song_id, artist_id) 
			VALUES ($1, $2)
			ON CONFLICT (song_id, artist_id) DO NOTHING
		`, id, artistId)
		if err != nil {
			return Song{}, err
		}
	}

	return Song{
		Id:             int(id),
		SpotifyId:      track.ID.String(),
		Name:           track.Name,
		ReleaseDate:    track.Album.ReleaseDateTime(),
		SpotifyAlbumId: track.Album.ID.String(),
		CreatedAt:      time.Now(),
	}, nil
}
