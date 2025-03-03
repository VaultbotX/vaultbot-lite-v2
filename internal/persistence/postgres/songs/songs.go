package songs

import (
	"github.com/jmoiron/sqlx"
	"github.com/zmb3/spotify/v2"
	"time"
)

type Song struct {
	Id               int       `db:"id"`
	SpotifyId        string    `db:"spotify_id"`
	Name             string    `db:"name"`
	ReleaseDate      time.Time `db:"release_date"`
	SpotifyAlbumId   string    `db:"spotify_album_id"`
	Acousticness     float32   `db:"acousticness"`
	Danceability     float32   `db:"danceability"`
	DurationMs       int       `db:"duration_ms"`
	Energy           float32   `db:"energy"`
	Instrumentalness float32   `db:"instrumentalness"`
	Key              int       `db:"key"`
	Liveness         float32   `db:"liveness"`
	Loudness         float32   `db:"loudness"`
	Mode             int       `db:"mode"`
	Speechiness      float32   `db:"speechiness"`
	Tempo            float32   `db:"tempo"`
	TimeSignature    int       `db:"time_signature"`
	Valence          float32   `db:"valence"`
	CreatedAt        time.Time `db:"created_at"`
}

// AddSong adds a song to the database
func AddSong(db *sqlx.DB, track *spotify.FullTrack, audioFeatures []*spotify.AudioFeatures) (Song, error) {
	firstAudioFeature := audioFeatures[0]
	row, err := db.NamedExec(`
		INSERT INTO songs (spotify_id, name, release_date, spotify_album_id, acousticness, danceability, duration_ms, energy, instrumentalness, key, liveness, loudness, mode, speechiness, tempo, time_signature, valence) 
		VALUES (:spotify_id, :name, :release_date, :spotify_album_id, :acousticness, :danceability, :duration_ms, :energy, :instrumentalness, :key, :liveness, :loudness, :mode, :speechiness, :tempo, :time_signature, :valence)
		ON CONFLICT (spotify_id) DO NOTHING
	`, map[string]any{
		"spotify_id":       track.ID.String(),
		"name":             track.Name,
		"release_date":     track.Album.ReleaseDate,
		"spotify_album_id": track.Album.ID.String(),
		"acousticness":     firstAudioFeature.Acousticness,
		"danceability":     firstAudioFeature.Danceability,
		"duration_ms":      firstAudioFeature.Duration,
		"energy":           firstAudioFeature.Energy,
		"instrumentalness": firstAudioFeature.Instrumentalness,
		"key":              firstAudioFeature.Key,
		"liveness":         firstAudioFeature.Liveness,
		"loudness":         firstAudioFeature.Loudness,
		"mode":             firstAudioFeature.Mode,
		"speechiness":      firstAudioFeature.Speechiness,
		"tempo":            firstAudioFeature.Tempo,
		"time_signature":   firstAudioFeature.TimeSignature,
		"valence":          firstAudioFeature.Valence,
	})

	if err != nil {
		return Song{}, err
	}

	id, err := row.LastInsertId()
	if err != nil {
		return Song{}, err
	}

	return Song{
		Id:               int(id),
		SpotifyId:        track.ID.String(),
		Name:             track.Name,
		ReleaseDate:      track.Album.ReleaseDateTime(),
		SpotifyAlbumId:   track.Album.ID.String(),
		Acousticness:     firstAudioFeature.Acousticness,
		Danceability:     firstAudioFeature.Danceability,
		DurationMs:       int(firstAudioFeature.Duration),
		Energy:           firstAudioFeature.Energy,
		Instrumentalness: firstAudioFeature.Instrumentalness,
		Key:              int(firstAudioFeature.Key),
		Liveness:         firstAudioFeature.Liveness,
		Loudness:         firstAudioFeature.Loudness,
		Mode:             int(firstAudioFeature.Mode),
		Speechiness:      firstAudioFeature.Speechiness,
		Tempo:            firstAudioFeature.Tempo,
		TimeSignature:    int(firstAudioFeature.TimeSignature),
		Valence:          firstAudioFeature.Valence,
		CreatedAt:        time.Now(),
	}, nil
}
