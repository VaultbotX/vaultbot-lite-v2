package songs

import "time"

type Song struct {
	Id               int       `db:"id"`
	SpotifyId        string    `db:"spotify_id"`
	Name             string    `db:"name"`
	ReleaseDate      time.Time `db:"release_date"`
	SpotifyAlbumId   string    `db:"spotify_album_id"`
	Acousticness     float64   `db:"acousticness"`
	Danceability     float64   `db:"danceability"`
	DurationMs       int       `db:"duration_ms"`
	Energy           float64   `db:"energy"`
	Instrumentalness float64   `db:"instrumentalness"`
	Key              int       `db:"key"`
	Liveness         float64   `db:"liveness"`
	Loudness         float64   `db:"loudness"`
	Mode             int       `db:"mode"`
	Speechiness      float64   `db:"speechiness"`
	Tempo            float64   `db:"tempo"`
	TimeSignature    int       `db:"time_signature"`
	Valence          float64   `db:"valence"`
	CreatedAt        time.Time `db:"created_at"`
}
