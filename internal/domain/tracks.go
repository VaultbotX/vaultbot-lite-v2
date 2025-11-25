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
	GetRandomGenreTracks() ([]songs.Song, error)
}

type AddTrackService struct {
	Repo AddTrackRepository
}

func NewTrackService(repo AddTrackRepository) *AddTrackService {
	return &AddTrackService{
		Repo: repo,
	}
}
