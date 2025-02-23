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

var DefaultGenre = Genre{}

// AddGenre adds a genre to the database
func AddGenre(db *sqlx.DB, name string) (Genre, error) {
	row, err := db.NamedExec(`
		INSERT INTO genres (name) 
		VALUES (:name)
		ON CONFLICT (name) DO NOTHING
	`, map[string]interface{}{
		"name": name,
	})

	if err != nil {
		return DefaultGenre, err
	}

	id, err := row.LastInsertId()
	if err != nil {
		return DefaultGenre, err
	}

	return Genre{
		Id:        int(id),
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}
