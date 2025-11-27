package cron

import (
	"context"
	"reflect"
	"testing"

	"time"

	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	psongs "github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	"github.com/zmb3/spotify/v2"
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

func makePI(id string) zspotify.PlaylistItem {
	return zspotify.PlaylistItem{AddedAt: time.Now().Format(spotify.TimestampLayout), Track: zspotify.PlaylistItemTrack{Track: &zspotify.FullTrack{SimpleTrack: zspotify.SimpleTrack{ID: zspotify.ID(id)}}}}
}

func TestPopulatePlaylistWithDeps_E2E(t *testing.T) {
	// current playlist has 1 and 2
	mockRepo := &mockPlaylistRepo{getPlaylistTracksResponse: []zspotify.PlaylistItem{makePI("1"), makePI("2")}}
	// desired rows are 2,3,4 -> should remove 1, add 3,4, keep 2
	mockTrack := &mockTrackRepo{rows: []psongs.Song{{SpotifyId: "2"}, {SpotifyId: "3"}, {SpotifyId: "4"}}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populatePlaylist(ctx, service, mockTrack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// verify removed contains 1
	if !containsID(mockRepo.removed, zspotify.ID("1")) || len(mockRepo.removed) != 1 {
		t.Fatalf("unexpected removed: %v", mockRepo.removed)
	}

	// verify added contains 3 and 4 in order
	if !reflect.DeepEqual(mockRepo.added, []zspotify.ID{zspotify.ID("3"), zspotify.ID("4")}) {
		t.Fatalf("unexpected added order: %v", mockRepo.added)
	}
}

func containsID(slice []zspotify.ID, id zspotify.ID) bool {
	for _, s := range slice {
		if s == id {
			return true
		}
	}
	return false
}
