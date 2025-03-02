package domain

import (
	"github.com/zmb3/spotify/v2"
	"time"
)

type CacheTrack struct {
	TrackId spotify.ID
	AddedAt time.Time
}

type TrackRepository interface {
	AddTrackToDatabase(fields *UserFields, track *spotify.FullTrack, artist []*spotify.FullArtist, audioFeatures *spotify.AudioFeatures) error
}

type TrackService struct {
	Repo TrackRepository
}

func NewTrackService(repo TrackRepository) *TrackService {
	return &TrackService{
		Repo: repo,
	}
}
