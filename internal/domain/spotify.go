package domain

import (
	"context"
	"github.com/zmb3/spotify/v2"
)

type EntityType int

const (
	Track EntityType = iota
	Artist
	Genre
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
	GetPlaylistTracks(playlistItemChan chan<- *spotify.PlaylistItem, ctx context.Context) error
	AddTracksToPlaylist(ctx context.Context, trackIds []spotify.ID) error
	RemoveTracksFromPlaylist(ctx context.Context, trackIds []spotify.ID) error
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
