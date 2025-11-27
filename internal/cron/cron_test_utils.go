package cron

import (
	"context"
	"time"

	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	psongs "github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	zspotify "github.com/zmb3/spotify/v2"
)

// mockPlaylistRepo implements the SpotifyPlaylistRepository used by domain.NewSpotifyPlaylistService
type mockPlaylistRepo struct {
	getPlaylistTracksResponse []zspotify.PlaylistItem
	removed                   []zspotify.ID
	added                     []zspotify.ID
	updatedDescription        string
}

func (m *mockPlaylistRepo) GetPlaylistTracks(ctx context.Context) ([]zspotify.PlaylistItem, error) {
	return m.getPlaylistTracksResponse, nil
}

func (m *mockPlaylistRepo) AddTracksToPlaylist(ctx context.Context, trackIds []zspotify.ID) error {
	m.added = append(m.added, trackIds...)
	return nil
}

func (m *mockPlaylistRepo) RemoveTracksFromPlaylist(ctx context.Context, trackIds []zspotify.ID) error {
	m.removed = append(m.removed, trackIds...)
	return nil
}

func (m *mockPlaylistRepo) UpdatePlaylistDescription(ctx context.Context, description string) error {
	m.updatedDescription = description
	return nil
}

// mockTrackRepo implements domain.AddTrackRepository
type mockTrackRepo struct {
	rows []psongs.Song
}

func (m *mockTrackRepo) AddTrackToDatabase(fields *domain.UserFields, track *zspotify.FullTrack, artist []*zspotify.FullArtist) error {
	panic("not implemented")
}
func (m *mockTrackRepo) GetRandomGenreTracks() (songs []psongs.Song, genreName string, err error) {
	return m.rows, "genre", nil
}
func (m *mockTrackRepo) GetTop50Tracks() (songs []psongs.Song, err error) {
	return m.rows, nil
}

func makePI(id string) zspotify.PlaylistItem {
	return zspotify.PlaylistItem{AddedAt: time.Now().Format(zspotify.TimestampLayout), Track: zspotify.PlaylistItemTrack{Track: &zspotify.FullTrack{SimpleTrack: zspotify.SimpleTrack{ID: zspotify.ID(id)}}}}
}

func containsID(slice []zspotify.ID, id zspotify.ID) bool {
	for _, s := range slice {
		if s == id {
			return true
		}
	}
	return false
}
