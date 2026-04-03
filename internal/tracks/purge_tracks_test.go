package tracks

import (
	"context"
	"testing"
	"time"

	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/zmb3/spotify/v2"
)

type mockPlaylistService struct {
	items []spotify.PlaylistItem
}

func (m *mockPlaylistService) GetPlaylistTracks(_ context.Context) ([]spotify.PlaylistItem, error) {
	return m.items, nil
}

func (m *mockPlaylistService) AddTracksToPlaylist(_ context.Context, _ []spotify.ID) error {
	return nil
}

func (m *mockPlaylistService) RemoveTracksFromPlaylist(_ context.Context, _ []spotify.ID) error {
	return nil
}

func (m *mockPlaylistService) UpdatePlaylistDescription(_ context.Context, _ string) error {
	return nil
}

func makeItem(id string, addedAt time.Time) spotify.PlaylistItem {
	return spotify.PlaylistItem{
		AddedAt: addedAt.UTC().Format(spotify.TimestampLayout),
		Track: spotify.PlaylistItemTrack{
			Track: &spotify.FullTrack{
				SimpleTrack: spotify.SimpleTrack{ID: spotify.ID(id)},
			},
		},
	}
}

func TestPurgeTracks_RemovesExpiredTracks(t *testing.T) {
	now := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		items    []spotify.PlaylistItem
		expected int
	}{
		{
			name: "Removes track added 15 days ago",
			items: []spotify.PlaylistItem{
				makeItem("old", now.Add(-15*24*time.Hour)),
				makeItem("recent", now.Add(-2*24*time.Hour)),
			},
			expected: 1,
		},
		{
			name: "Removes nothing when all tracks are recent",
			items: []spotify.PlaylistItem{
				makeItem("a", now.Add(-1*24*time.Hour)),
				makeItem("b", now.Add(-7*24*time.Hour)),
			},
			expected: 0,
		},
		{
			name: "Removes all tracks when all are expired",
			items: []spotify.PlaylistItem{
				makeItem("x", now.Add(-15*24*time.Hour)),
				makeItem("y", now.Add(-20*24*time.Hour)),
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := domain.NewSpotifyPlaylistService(&mockPlaylistService{items: tt.items})
			numRemoved, err := PurgeTracks(context.Background(), now, svc)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if numRemoved != tt.expected {
				t.Errorf("Expected %d removed, got %d", tt.expected, numRemoved)
			}
		})
	}
}
