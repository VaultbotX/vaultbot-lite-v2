package archive

import "time"

type Archive struct {
	Id        int       `db:"id"`
	SongId    int       `db:"song_id"`
	UserId    int       `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}
