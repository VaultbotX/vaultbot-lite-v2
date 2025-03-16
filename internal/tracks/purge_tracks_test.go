package tracks

import (
	"context"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"testing"
	"time"
)

func TestPurgeTracks_RemovesExpiredTracks(t *testing.T) {
	// Arrange
	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	existingTracks := []*domain.CacheTrack{
		// Added on 2024-01-13
		{
			TrackId: "1",
			AddedAt: date.Add(-2 * 24 * time.Hour),
		},
		// Added 8 days prior to 2025-01-15
		{
			TrackId: "2",
			AddedAt: date.Add(-8 * 24 * time.Hour),
		},
		// Added 15 days prior to 2025-01-15
		{
			TrackId: "3",
			AddedAt: date.Add(-15 * 24 * time.Hour),
		},
	}

	tests := []struct {
		name            string
		purgeDaysInPast int64
		expected        int
	}{
		{"Purges tracks older than 1 day", 1, 3},
		{"Purges tracks older than 1 week", 7, 2},
		{"Purges tracks older than 2 weeks", 14, 1},
		{"Purges tracks older than 3 weeks", 21, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			persistence.TrackCache = persistence.NewCache()
			persistence.TrackCache.SetMulti(existingTracks)

			purgeDaysInPastMillis := tt.purgeDaysInPast * 24 * time.Hour.Milliseconds()
			preferenceService := domain.NewPreferenceService(&mockPreferenceRepo{
				preferences: map[domain.PreferenceKey]domain.Preference{
					domain.MaxDurationKey: {
						Key:   domain.MaxTrackAgeKey,
						Value: toBytes(purgeDaysInPastMillis),
					},
				},
			})

			spotifyPlaylistService := domain.NewSpotifyPlaylistService(&mockPlaylistService{})

			// Act
			numRemoved, err := PurgeTracks(context.Background(), date, preferenceService, spotifyPlaylistService)

			// Assert
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if numRemoved != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, numRemoved)
			}
		})
	}
}
