package genres

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Genre struct {
	Id        int       `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

// AddGenre adds a genre to the database
func AddGenre(tx *sqlx.Tx, name string) (Genre, error) {
	var addGenre Genre

	err := tx.QueryRowx(`
		SELECT id, name, created_at
		FROM genres
		WHERE name = $1
	`, name).StructScan(&addGenre)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return Genre{}, err
		}

		// no genre found, create a new one
		err := tx.QueryRowx(`
			INSERT INTO genres (name) 
			VALUES ($1)
			ON CONFLICT (name) DO NOTHING
			RETURNING id, created_at
		`, name).StructScan(&addGenre)

		if err != nil {
			return Genre{}, err
		}

		addGenre.Name = name
	}

	return addGenre, nil
}

// GetRandomGenre retrieves a random genre with more than 20 associated songs
func GetRandomGenre(db *sqlx.DB) (Genre, error) {
	var genre Genre

	err := db.QueryRowx(`
		SELECT g.id, g.name, g.created_at, COUNT(lsg.song_id) AS count
		FROM genres g
				 JOIN link_song_genres lsg ON g.id = lsg.genre_id
		GROUP BY g.id
		HAVING COUNT(lsg.song_id) > 20
		ORDER BY RANDOM()
		LIMIT 1;
	`).StructScan(&genre)

	if err != nil {
		return Genre{}, err
	}

	return genre, nil
}
