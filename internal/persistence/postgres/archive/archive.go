package archive

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Archive struct {
	Id        int       `db:"id"`
	SongId    int       `db:"song_id"`
	CreatedAt time.Time `db:"created_at"`
}

// AddArchive adds an archive entry to the database
func AddArchive(tx *sqlx.Tx, songId int) (Archive, error) {
	var addArchive Archive
	err := tx.QueryRowx(`
		INSERT INTO song_archive (song_id)
		VALUES ($1)
		RETURNING id, created_at
	`, songId).StructScan(&addArchive)

	if err != nil {
		return Archive{}, err
	}

	addArchive.SongId = songId

	return addArchive, nil
}
