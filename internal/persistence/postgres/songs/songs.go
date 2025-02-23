package songs

import (
	"github.com/jmoiron/sqlx"
	"time"
)

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

var DefaultSong = Song{}

// AddSong adds a song to the database
func AddSong(db *sqlx.DB, spotifyId string, name string, releaseDate time.Time, spotifyAlbumId string, acousticness float64, danceability float64, durationMs int, energy float64, instrumentalness float64, key int, liveness float64, loudness float64, mode int, speechiness float64, tempo float64, timeSignature int, valence float64) (Song, error) {
	row, err := db.NamedExec(`
		INSERT INTO songs (spotify_id, name, release_date, spotify_album_id, acousticness, danceability, duration_ms, energy, instrumentalness, key, liveness, loudness, mode, speechiness, tempo, time_signature, valence) 
		VALUES (:spotify_id, :name, :release_date, :spotify_album_id, :acousticness, :danceability, :duration_ms, :energy, :instrumentalness, :key, :liveness, :loudness, :mode, :speechiness, :tempo, :time_signature, :valence)
		ON CONFLICT (spotify_id) DO NOTHING
	`, map[string]interface{}{
		"spotify_id":       spotifyId,
		"name":             name,
		"release_date":     releaseDate,
		"spotify_album_id": spotifyAlbumId,
		"acousticness":     acousticness,
		"danceability":     danceability,
		"duration_ms":      durationMs,
		"energy":           energy,
		"instrumentalness": instrumentalness,
		"key":              key,
		"liveness":         liveness,
		"loudness":         loudness,
		"mode":             mode,
		"speechiness":      speechiness,
		"tempo":            tempo,
		"time_signature":   timeSignature,
		"valence":          valence,
	})

	if err != nil {
		return DefaultSong, err
	}

	id, err := row.LastInsertId()
	if err != nil {
		return DefaultSong, err
	}

	return Song{
		Id:               int(id),
		SpotifyId:        spotifyId,
		Name:             name,
		ReleaseDate:      releaseDate,
		SpotifyAlbumId:   spotifyAlbumId,
		Acousticness:     acousticness,
		Danceability:     danceability,
		DurationMs:       durationMs,
		Energy:           energy,
		Instrumentalness: instrumentalness,
		Key:              key,
		Liveness:         liveness,
		Loudness:         loudness,
		Mode:             mode,
		Speechiness:      speechiness,
		Tempo:            tempo,
		TimeSignature:    timeSignature,
		Valence:          valence,
		CreatedAt:        time.Now(),
	}, nil
}
