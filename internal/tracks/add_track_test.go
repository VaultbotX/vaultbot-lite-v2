package tracks

import (
	"context"
	"errors"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/zmb3/spotify/v2"
	"testing"
)

/*
Cases:
1. Invalid track ID
2. Track already in playlist
3. Track not found
4. Artist blacklisted
5. Track blacklisted
6. Track too long
7. Happy path
*/

func TestAddTrack_ShortCircuits_WithInvalidSongId(t *testing.T) {
	// Arrange
	input := &AddTrackInput{
		TrackId: "invalid-track-id",
	}

	// Act
	_, err := AddTrack(input)

	// Assert
	if !errors.Is(err, domain.ErrInvalidSpotifyId) {
		t.Errorf("Expected %v, got %v", domain.ErrInvalidSpotifyId, err)
	}
}

func TestAddTrack_ShortCircuits_IfTrackIsInPlaylist(t *testing.T) {
	// Arrange
	input := &AddTrackInput{
		TrackId: "validtrackid",
	}
	persistence.TrackCache.Set(&domain.CacheTrack{
		TrackId: "validtrackid",
	})

	// Act
	_, err := AddTrack(input)

	// Assert
	if !errors.Is(err, domain.ErrTrackAlreadyInPlaylist) {
		t.Errorf("Expected %v, got %v", domain.ErrTrackAlreadyInPlaylist, err)
	}
}

type MockSpotifyTrackRepo struct {
	getTrackResponse *spotify.FullTrack
}

func (m *MockSpotifyTrackRepo) GetTrack(_ spotify.ID, trackChan chan<- *spotify.FullTrack, _ context.Context) error {
	trackChan <- m.getTrackResponse
	return nil
}

func TestAddTrack_ShortCircuits_IfTrackNotFound(t *testing.T) {
	// Arrange
	spotifyTrackService := domain.NewSpotifyTrackService(&MockSpotifyTrackRepo{
		getTrackResponse: nil,
	})
	input := &AddTrackInput{
		TrackId:        "validtrackid2",
		SpTrackService: spotifyTrackService,
	}

	// Act
	_, err := AddTrack(input)

	// Assert
	if !errors.Is(err, domain.ErrNoTrackExists) {
		t.Errorf("Expected %v, got %v", domain.ErrNoTrackExists, err)
	}
}
