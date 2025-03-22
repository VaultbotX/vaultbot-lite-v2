package domain

import (
	"github.com/zmb3/spotify/v2"
	"time"
)

type CacheTrack struct {
	TrackId spotify.ID
	AddedAt time.Time
}

type AddTrackRepository interface {
	AddTrackToDatabase(fields *UserFields, track *spotify.FullTrack, artist []*spotify.FullArtist) error
}

type AddTrackService struct {
	Repo AddTrackRepository
}

func NewTrackService(repo AddTrackRepository) *AddTrackService {
	return &AddTrackService{
		Repo: repo,
	}
}
