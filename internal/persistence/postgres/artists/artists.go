package artists

import "time"

type Artist struct {
	Id        int       `db:"id"`
	SpotifyId string    `db:"spotify_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}
