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
func AddArtist(db *sqlx.DB, spotifyId string, name string) (Artist, error) {
	row, err := db.NamedExec(`
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

	return Artist{
		Id:        int(id),
		SpotifyId: spotifyId,
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}
