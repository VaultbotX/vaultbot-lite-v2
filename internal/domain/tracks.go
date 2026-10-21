package domain

import (
	"time"

	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	"github.com/zmb3/spotify/v2"
)

type AddTrackRepository interface {
	AddTrackToDatabase(track *spotify.FullTrack, artists []*spotify.FullArtist) error
	GetRandomGenreTracks() (songs []songs.Song, genreName string, err error)
	GetTop50Tracks() (songs []songs.Song, err error)
	GetTopWindowTracks(windowDuration, bufferDuration time.Duration, minCount, limit, maxAttempts int) (songs []songs.Song, start, end time.Time, err error)
	GetRandomTracks(limit int) (songs []songs.Song, err error)
}

type AddTrackService struct {
	Repo AddTrackRepository
}

func NewTrackService(repo AddTrackRepository) *AddTrackService {
	return &AddTrackService{
		Repo: repo,
	}
}
