package genres

import (
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
	row, err := tx.NamedExec(`
		INSERT INTO genres (name) 
		VALUES (:name)
		ON CONFLICT (name) DO NOTHING
	`, map[string]any{
		"name": name,
	})

	if err != nil {
		return Genre{}, err
	}

	id, err := row.LastInsertId()
	if err != nil {
		return Genre{}, err
	}

	return Genre{
		Id:        int(id),
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}
