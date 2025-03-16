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

type mockSpotifyTrackRepo struct {
	getTrackResponse *spotify.FullTrack
}

func (m *mockSpotifyTrackRepo) GetTrack(_ spotify.ID, trackChan chan<- *spotify.FullTrack, _ context.Context) error {
	trackChan <- m.getTrackResponse
	return nil
}

type mockSpotifyArtistRepo struct {
	getArtistsResponse []*spotify.FullArtist
}

func (m *mockSpotifyArtistRepo) GetArtists(_ []spotify.ID, artistChan chan<- *spotify.FullArtist, _ context.Context) error {
	for _, artist := range m.getArtistsResponse {
		artistChan <- artist
	}
	return nil
}

func TestAddTrack_ShortCircuits_IfTrackNotFound(t *testing.T) {
	// Arrange
	spotifyTrackService := domain.NewSpotifyTrackService(&mockSpotifyTrackRepo{
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

type blacklistedItem struct {
	blacklistType domain.EntityType
	id            string
}

type mockBlacklistRepo struct {
	blacklistedItems []blacklistedItem
}

func (m *mockBlacklistRepo) AddToBlacklist(_ context.Context, _ domain.EntityType, _ string,
	_ *domain.UserFields) error {
	return nil
}

func (m *mockBlacklistRepo) RemoveFromBlacklist(_ context.Context, _ domain.EntityType, _ string) error {
	return nil
}

func (m *mockBlacklistRepo) CheckBlacklistItem(_ context.Context, blacklistType domain.EntityType, id string) (bool, error) {
	for _, item := range m.blacklistedItems {
		if item.blacklistType == blacklistType && item.id == id {
			return true, nil
		}
	}

	return false, nil
}

func TestAddTrack_ShortCircuits_IfArtistBlacklisted(t *testing.T) {
	// Arrange
	spotifyTrackService := domain.NewSpotifyTrackService(&mockSpotifyTrackRepo{
		getTrackResponse: &spotify.FullTrack{
			SimpleTrack: spotify.SimpleTrack{
				Artists: []spotify.SimpleArtist{
					{
						ID:   "blacklistedartistid",
						Name: "Blacklisted Artist",
					},
				},
			},
		},
	})
	spotifyArtistService := domain.NewSpotifyArtistService(&mockSpotifyArtistRepo{
		getArtistsResponse: []*spotify.FullArtist{
			{
				SimpleArtist: spotify.SimpleArtist{
					ID:   "blacklistedartistid",
					Name: "Blacklisted Artist",
				},
			},
		},
	})
	blacklistService := domain.NewBlacklistService(&mockBlacklistRepo{
		blacklistedItems: []blacklistedItem{
			{
				blacklistType: domain.Artist,
				id:            "blacklistedartistid",
			},
		},
	})
	input := &AddTrackInput{
		TrackId:          "validtrackid3",
		SpTrackService:   spotifyTrackService,
		SpArtistService:  spotifyArtistService,
		BlacklistService: blacklistService,
	}

	// Act
	_, err := AddTrack(input)

	// Assert
	var eab *domain.ErrArtistBlacklisted
	ok := errors.As(err, &eab)
	if !ok {
		t.Errorf("Expected %v, got %v", &domain.ErrArtistBlacklisted{}, err)
	}
}

func TestAddTrack_ShortCircuits_IfTrackBlacklisted(t *testing.T) {
	// Arrange
	spotifyTrackService := domain.NewSpotifyTrackService(&mockSpotifyTrackRepo{
		getTrackResponse: &spotify.FullTrack{
			SimpleTrack: spotify.SimpleTrack{
				ID: "blacklistedtrackid",
			},
		},
	})
	blacklistService := domain.NewBlacklistService(&mockBlacklistRepo{
		blacklistedItems: []blacklistedItem{
			{
				blacklistType: domain.Track,
				id:            "blacklistedtrackid",
			},
		},
	})
	input := &AddTrackInput{
		TrackId:          "validtrackid4",
		SpTrackService:   spotifyTrackService,
		BlacklistService: blacklistService,
	}

	// Act
	_, err := AddTrack(input)

	// Assert
	var etb *domain.ErrTrackBlacklisted
	ok := errors.As(err, &etb)
	if !ok {
		t.Errorf("Expected %v, got %v", &domain.ErrTrackBlacklisted{}, err)
	}
}
