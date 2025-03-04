package artists

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"time"
)

type Artist struct {
	Id        int       `db:"id"`
	SpotifyId string    `db:"spotify_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

// AddArtist adds an artist to the database
func AddArtist(tx *sqlx.Tx, spotifyId string, name string, genreIds []int) (Artist, error) {
	var addArtist Artist

	err := tx.QueryRowx(`
		SELECT id, spotify_id, name, created_at
		FROM artists
		WHERE spotify_id = $1
	`, spotifyId).StructScan(&addArtist)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return Artist{}, err
		}

		err := tx.QueryRowx(`
			INSERT INTO artists (spotify_id, name) 
			VALUES ($1, $2)
			ON CONFLICT (spotify_id) DO NOTHING
			RETURNING id, created_at
		`, spotifyId, name).StructScan(&addArtist)

		if err != nil {
			return Artist{}, err
		}

		addArtist.SpotifyId = spotifyId
		addArtist.Name = name
	}

	// insert link records for genres
	for _, genreId := range genreIds {
		_, err := tx.Exec(`
			INSERT INTO link_artist_genres (artist_id, genre_id) 
			VALUES ($1, $2)
			ON CONFLICT (artist_id, genre_id) DO NOTHING
		`, addArtist.Id, genreId)
		if err != nil {
			return Artist{}, err
		}
	}

	return addArtist, nil
}
