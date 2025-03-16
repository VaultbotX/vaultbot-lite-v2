package tracks

import (
	"context"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/zmb3/spotify/v2"
	"testing"
	"time"
)

func getPlaylistItem(id string, date time.Time) *spotify.PlaylistItem {
	return &spotify.PlaylistItem{
		AddedAt: date.Format(spotify.TimestampLayout),
		Track: spotify.PlaylistItemTrack{
			Track: &spotify.FullTrack{
				SimpleTrack: spotify.SimpleTrack{
					ID: spotify.ID(id),
				},
			},
			Episode: nil,
		},
	}
}

func TestCacheTracks_CachesTracks(t *testing.T) {
	// Arrange
	persistence.TrackCache = persistence.NewCache()
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	tracks := []*spotify.PlaylistItem{
		getPlaylistItem("1", date),
		getPlaylistItem("2", date.Add(-1*time.Hour)),
		getPlaylistItem("3", date.Add(-2*time.Hour)),
		getPlaylistItem("4", date.Add(-3*time.Hour)),
	}
	playlistService := domain.NewSpotifyPlaylistService(&mockPlaylistService{
		getPlaylistTracksResponse: tracks,
	})

	// Act
	err := CacheTracks(context.Background(), playlistService)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(persistence.TrackCache.GetAll()) != len(tracks) {
		t.Errorf("Expected %d tracks in cache, got %d", len(tracks), len(persistence.TrackCache.GetAll()))
	}
}

func TestCacheTracks_NoOp_WhenEmptyPlaylist(t *testing.T) {
	// Arrange
	persistence.TrackCache = persistence.NewCache()
	playlistService := domain.NewSpotifyPlaylistService(&mockPlaylistService{
		getPlaylistTracksResponse: []*spotify.PlaylistItem{},
	})

	// Act
	err := CacheTracks(context.Background(), playlistService)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(persistence.TrackCache.GetAll()) != 0 {
		t.Errorf("Expected no tracks in cache, got %d", len(persistence.TrackCache.GetAll()))
	}
}
