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

type TrackPartial struct {
	Id        int
	SpotifyId spotify.ID
}

type TrackPartialWithMetadata struct {
	TrackPartial
	Length      time.Duration
	AlbumId     spotify.ID
	ReleaseDate time.Time
}

type DuplicateTrackCheckingRepository interface {
	GetRelatedTracks(trackId spotify.ID) ([]TrackPartialWithMetadata, error)
	SetDuplicateTrack(sourceTrackId spotify.ID, targetTrackId spotify.ID) error
}
