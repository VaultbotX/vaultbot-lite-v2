package domain

import (
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
)

type TrackRepository interface {
	AddTrackToDatabase(fields *types.UserFields, track *spotify.FullTrack, artist []*spotify.FullArtist, audioFeatures *spotify.AudioFeatures) error
}

type TrackService struct {
	Repo TrackRepository
}

func NewTrackService(repo TrackRepository) *TrackService {
	return &TrackService{
		Repo: repo,
	}
}
