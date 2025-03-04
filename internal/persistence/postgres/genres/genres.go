package genres

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"time"
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
