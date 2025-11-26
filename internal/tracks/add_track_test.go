package tracks

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	"github.com/zmb3/spotify/v2"
)

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
	close(trackChan)
	return nil
}

type mockSpotifyArtistRepo struct {
	getArtistsResponse []*spotify.FullArtist
}

func (m *mockSpotifyArtistRepo) GetArtists(_ []spotify.ID, artistChan chan<- *spotify.FullArtist, _ context.Context) error {
	for _, artist := range m.getArtistsResponse {
		artistChan <- artist
	}
	close(artistChan)
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

type mockPreferenceRepo struct {
	preferences map[domain.PreferenceKey]domain.Preference
}

func (m *mockPreferenceRepo) Set(_ context.Context, _ domain.PreferenceKey, _ any) error {
	return nil
}

func (m *mockPreferenceRepo) Get(_ context.Context, _ domain.PreferenceKey) (*domain.Preference, error) {
	if pref, ok := m.preferences[domain.MaxDurationKey]; ok {
		return &pref, nil
	}

	return nil, nil
}

func (m *mockPreferenceRepo) GetAll(_ context.Context) (map[domain.PreferenceKey]domain.Preference, error) {
	return m.preferences, nil
}

func toBytes(value int64) []byte {
	b, err := json.Marshal(value)
	if err != nil {
		return nil
	}

	return b
}

func TestAddTrack_ShortCircuits_IfTrackTooLong(t *testing.T) {
	// Arrange
	spotifyTrackService := domain.NewSpotifyTrackService(&mockSpotifyTrackRepo{
		getTrackResponse: &spotify.FullTrack{
			SimpleTrack: spotify.SimpleTrack{
				ID:       "toolongtrackid",
				Duration: 600_001, // 10 minutes and 1 millisecond
			},
		},
	})
	blacklistService := domain.NewBlacklistService(&mockBlacklistRepo{
		blacklistedItems: []blacklistedItem{},
	})
	preferenceService := domain.NewPreferenceService(&mockPreferenceRepo{
		preferences: map[domain.PreferenceKey]domain.Preference{
			domain.MaxDurationKey: {
				Key:   domain.MaxDurationKey,
				Value: toBytes(600_000), // 10 minutes
			},
		},
	})

	input := &AddTrackInput{
		TrackId:           "toolongtrackid",
		SpTrackService:    spotifyTrackService,
		PreferenceService: preferenceService,
		BlacklistService:  blacklistService,
	}

	// Act
	_, err := AddTrack(input)

	// Assert
	if !errors.Is(err, domain.ErrTrackTooLong) {
		t.Errorf("Expected %v, got %v", domain.ErrTrackTooLong, err)
	}
}

func TestAddTrack_ShortCircuits_IfArtistGenreBlacklisted(t *testing.T) {
	// Arrange
	spotifyTrackService := domain.NewSpotifyTrackService(&mockSpotifyTrackRepo{
		getTrackResponse: &spotify.FullTrack{
			SimpleTrack: spotify.SimpleTrack{
				ID:       "trackid",
				Duration: 300_000,
			},
		},
	})
	spotifyArtistService := domain.NewSpotifyArtistService(&mockSpotifyArtistRepo{
		getArtistsResponse: []*spotify.FullArtist{
			{
				SimpleArtist: spotify.SimpleArtist{
					ID:   "artistid",
					Name: "Artist With Blacklisted Genre",
				},
				Genres: []string{"blacklistedgenre"},
			},
		},
	})
	blacklistService := domain.NewBlacklistService(&mockBlacklistRepo{
		blacklistedItems: []blacklistedItem{
			{
				blacklistType: domain.Genre,
				id:            "blacklistedgenre",
			},
		},
	})
	preferenceService := domain.NewPreferenceService(&mockPreferenceRepo{
		preferences: map[domain.PreferenceKey]domain.Preference{
			domain.MaxDurationKey: {
				Key:   domain.MaxDurationKey,
				Value: toBytes(600_000), // 10 minutes
			},
		},
	})
	input := &AddTrackInput{
		TrackId:           "trackid",
		SpTrackService:    spotifyTrackService,
		SpArtistService:   spotifyArtistService,
		BlacklistService:  blacklistService,
		PreferenceService: preferenceService,
	}

	// Act
	_, err := AddTrack(input)

	// Assert
	var etb *domain.ErrGenreBlacklisted
	ok := errors.As(err, &etb)
	if !ok {
		t.Errorf("Expected %v, got %v", &domain.ErrGenreBlacklisted{}, err)
	}
}

type mockPlaylistService struct {
	getPlaylistTracksResponse []*spotify.PlaylistItem
}

func (m *mockPlaylistService) GetPlaylistTracks(_ context.Context) ([]spotify.PlaylistItem, error) {
	// Convert stored pointers to a slice of values to match the current interface.
	result := make([]spotify.PlaylistItem, 0, len(m.getPlaylistTracksResponse))
	for _, p := range m.getPlaylistTracksResponse {
		if p == nil {
			// append zero value if nil
			result = append(result, spotify.PlaylistItem{})
		} else {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (m *mockPlaylistService) AddTracksToPlaylist(_ context.Context, _ []spotify.ID) error {
	return nil
}

func (m *mockPlaylistService) RemoveTracksFromPlaylist(_ context.Context, _ []spotify.ID) error {
	return nil
}

type mockTrackRepository struct{}

func (m *mockTrackRepository) AddTrackToDatabase(_ *domain.UserFields, _ *spotify.FullTrack, _ []*spotify.FullArtist) error {
	return nil
}

func (m *mockTrackRepository) GetRandomGenreTracks() (songs []songs.Song, genreName string, err error) {
	return nil, "", nil
}

func TestAddTrack_ReturnsTrack_WhenValidTrack(t *testing.T) {
	// Arrange
	spotifyTrackService := domain.NewSpotifyTrackService(&mockSpotifyTrackRepo{
		getTrackResponse: &spotify.FullTrack{
			SimpleTrack: spotify.SimpleTrack{
				ID:       "validtrackidhappy",
				Duration: 300_000, // 5 minutes
			},
		},
	})
	spotifyArtistService := domain.NewSpotifyArtistService(&mockSpotifyArtistRepo{
		getArtistsResponse: []*spotify.FullArtist{
			{
				SimpleArtist: spotify.SimpleArtist{
					ID:   "artistidhappy",
					Name: "Artist Happy",
				},
			},
		},
	})
	blacklistService := domain.NewBlacklistService(&mockBlacklistRepo{
		blacklistedItems: []blacklistedItem{},
	})
	preferenceService := domain.NewPreferenceService(&mockPreferenceRepo{
		preferences: map[domain.PreferenceKey]domain.Preference{
			domain.MaxDurationKey: {
				Key:   domain.MaxDurationKey,
				Value: toBytes(600_000), // 10 minutes
			},
		},
	})
	playlistService := domain.NewSpotifyPlaylistService(&mockPlaylistService{})
	trackService := domain.NewTrackService(&mockTrackRepository{})

	input := &AddTrackInput{
		TrackId:           "validtrackidhappy",
		SpTrackService:    spotifyTrackService,
		SpArtistService:   spotifyArtistService,
		PreferenceService: preferenceService,
		BlacklistService:  blacklistService,
		SpPlaylistService: playlistService,
		TrackService:      trackService,
	}

	// Act
	result, err := AddTrack(input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Errorf("Expected a track, got nil")
	}
}
