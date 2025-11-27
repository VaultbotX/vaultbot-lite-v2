package domain

import (
	"time"

	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	"github.com/zmb3/spotify/v2"
)

type CacheTrack struct {
	TrackId spotify.ID
	AddedAt time.Time
}

type AddTrackRepository interface {
	AddTrackToDatabase(fields *UserFields, track *spotify.FullTrack, artist []*spotify.FullArtist) error
	GetRandomGenreTracks() (songs []songs.Song, genreName string, err error)
	GetTop50Tracks() (songs []songs.Song, err error)
}

type AddTrackService struct {
	Repo AddTrackRepository
}

func NewTrackService(repo AddTrackRepository) *AddTrackService {
	return &AddTrackService{
		Repo: repo,
	}
}
