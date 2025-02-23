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

var DefaultArchive = Archive{}

// AddArchive adds an archive to the database
func AddArchive(db *sqlx.DB, songId int, userId int) (Archive, error) {
	row, err := db.NamedExec(`
		INSERT INTO song_archive (song_id, user_id) 
		VALUES (:song_id, :user_id)
		ON CONFLICT (song_id, user_id) DO NOTHING
	`, map[string]interface{}{
		"song_id": songId,
		"user_id": userId,
	})

	if err != nil {
		return DefaultArchive, err
	}

	id, err := row.LastInsertId()
	if err != nil {
		return DefaultArchive, err
	}

	return Archive{
		Id:        int(id),
		SongId:    songId,
		UserId:    userId,
		CreatedAt: time.Now(),
	}, nil
}
