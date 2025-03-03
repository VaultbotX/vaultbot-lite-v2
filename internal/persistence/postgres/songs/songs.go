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
func AddSong(tx *sqlx.Tx, track *spotify.FullTrack, audioFeatures *spotify.AudioFeatures, genreIds []int, artistIds []int) (Song, error) {
	row, err := tx.NamedExec(`
		INSERT INTO songs (spotify_id, name, release_date, spotify_album_id, acousticness, danceability, duration_ms, energy, instrumentalness, key, liveness, loudness, mode, speechiness, tempo, time_signature, valence) 
		VALUES (:spotify_id, :name, :release_date, :spotify_album_id, :acousticness, :danceability, :duration_ms, :energy, :instrumentalness, :key, :liveness, :loudness, :mode, :speechiness, :tempo, :time_signature, :valence)
		ON CONFLICT (spotify_id) DO NOTHING
	`, map[string]any{
		"spotify_id":       track.ID.String(),
		"name":             track.Name,
		"release_date":     track.Album.ReleaseDate,
		"spotify_album_id": track.Album.ID.String(),
		"acousticness":     audioFeatures.Acousticness,
		"danceability":     audioFeatures.Danceability,
		"duration_ms":      audioFeatures.Duration,
		"energy":           audioFeatures.Energy,
		"instrumentalness": audioFeatures.Instrumentalness,
		"key":              audioFeatures.Key,
		"liveness":         audioFeatures.Liveness,
		"loudness":         audioFeatures.Loudness,
		"mode":             audioFeatures.Mode,
		"speechiness":      audioFeatures.Speechiness,
		"tempo":            audioFeatures.Tempo,
		"time_signature":   audioFeatures.TimeSignature,
		"valence":          audioFeatures.Valence,
	})

	if err != nil {
		return Song{}, err
	}

	id, err := row.LastInsertId()
	if err != nil {
		return Song{}, err
	}

	// insert link records
	for _, genreId := range genreIds {
		_, err := tx.Exec(`
			INSERT INTO link_song_genres (song_id, genre_id) 
			VALUES ($1, $2)
			ON CONFLICT (song_id, genre_id) DO NOTHING
		`, id, genreId)
		if err != nil {
			return Song{}, err
		}
	}

	for _, artistId := range artistIds {
		_, err := tx.Exec(`
			INSERT INTO link_song_artists (song_id, artist_id) 
			VALUES ($1, $2)
			ON CONFLICT (song_id, artist_id) DO NOTHING
		`, id, artistId)
		if err != nil {
			return Song{}, err
		}
	}

	return Song{
		Id:               int(id),
		SpotifyId:        track.ID.String(),
		Name:             track.Name,
		ReleaseDate:      track.Album.ReleaseDateTime(),
		SpotifyAlbumId:   track.Album.ID.String(),
		Acousticness:     audioFeatures.Acousticness,
		Danceability:     audioFeatures.Danceability,
		DurationMs:       int(audioFeatures.Duration),
		Energy:           audioFeatures.Energy,
		Instrumentalness: audioFeatures.Instrumentalness,
		Key:              int(audioFeatures.Key),
		Liveness:         audioFeatures.Liveness,
		Loudness:         audioFeatures.Loudness,
		Mode:             int(audioFeatures.Mode),
		Speechiness:      audioFeatures.Speechiness,
		Tempo:            audioFeatures.Tempo,
		TimeSignature:    int(audioFeatures.TimeSignature),
		Valence:          audioFeatures.Valence,
		CreatedAt:        time.Now(),
	}, nil
}
