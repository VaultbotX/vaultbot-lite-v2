package artists

import (
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
	row, err := tx.NamedExec(`
		INSERT INTO artists (spotify_id, name) 
		VALUES (:spotify_id, :name)
		ON CONFLICT (spotify_id) DO NOTHING
	`, map[string]any{
		"spotify_id": spotifyId,
		"name":       name,
	})

	if err != nil {
		return Artist{}, err
	}

	id, err := row.LastInsertId()
	if err != nil {
		return Artist{}, err
	}

	// insert link records for genres
	for _, genreId := range genreIds {
		_, err := tx.Exec(`
			INSERT INTO link_artist_genres (artist_id, genre_id) 
			VALUES ($1, $2)
			ON CONFLICT (artist_id, genre_id) DO NOTHING
		`, id, genreId)
		if err != nil {
			return Artist{}, err
		}
	}

	return Artist{
		Id:        int(id),
		SpotifyId: spotifyId,
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}
