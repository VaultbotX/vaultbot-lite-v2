package archive

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type Archive struct {
	Id        int       `db:"id"`
	SongId    int       `db:"song_id"`
	UserId    int       `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

// AddArchive adds an archive to the database
func AddArchive(tx *sqlx.Tx, songId int, userId int) (Archive, error) {
	var addArchive Archive
	err := tx.QueryRowx(`
		INSERT INTO song_archive (song_id, user_id) 
		VALUES ($1, $2)
		ON CONFLICT (song_id, user_id) DO NOTHING
	`, songId, userId).StructScan(&addArchive)

	if err != nil {
		return Archive{}, err
	}

	addArchive.SongId = songId
	addArchive.UserId = userId

	return addArchive, nil
}
