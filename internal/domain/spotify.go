package domain

import (
	"context"

	"github.com/zmb3/spotify/v2"
)

type Playlist int

const (
	// DynamicPlaylist represents the main playlist that is regularly updated with new tracks.
	DynamicPlaylist Playlist = iota
	// GenrePlaylist represents a playlist curated based on specific genres.
	GenrePlaylist
	// HighScoresPlaylist represents a playlist that features the tracks added most frequently.
	HighScoresPlaylist
	// ThrowbackPlaylist represents a playlist curated from the single release year with the most archived tracks.
	ThrowbackPlaylist
	// VarietyPlaylist represents a playlist of 100 randomly selected tracks from the archive.
	VarietyPlaylist
)

type SpotifyTrackRepository interface {
	GetTrack(trackId spotify.ID, trackChan chan<- *spotify.FullTrack, ctx context.Context) error
}

type SpotifyTrackService struct {
	Repo SpotifyTrackRepository
}

func NewSpotifyTrackService(repo SpotifyTrackRepository) *SpotifyTrackService {
	return &SpotifyTrackService{
		Repo: repo,
	}
}

type SpotifyPlaylistRepository interface {
	GetPlaylistTracks(ctx context.Context) ([]spotify.PlaylistItem, error)
	AddTracksToPlaylist(ctx context.Context, trackIds []spotify.ID) error
	RemoveTracksFromPlaylist(ctx context.Context, trackIds []spotify.ID) error
	UpdatePlaylistDescription(ctx context.Context, description string) error
}

type SpotifyPlaylistService struct {
	Repo SpotifyPlaylistRepository
}

func NewSpotifyPlaylistService(repo SpotifyPlaylistRepository) *SpotifyPlaylistService {
	return &SpotifyPlaylistService{
		Repo: repo,
	}
}

type SpotifyArtistRepository interface {
	GetArtists(artistIds []spotify.ID, artistChan chan<- *spotify.FullArtist, ctx context.Context) error
}

type SpotifyArtistService struct {
	Repo SpotifyArtistRepository
}

func NewSpotifyArtistService(repo SpotifyArtistRepository) *SpotifyArtistService {
	return &SpotifyArtistService{
		Repo: repo,
	}
}
