package cron

import (
	"context"
	"reflect"
	"testing"

	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	psongs "github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	zspotify "github.com/zmb3/spotify/v2"
)

func TestPopulateThrowbackPlaylist_AddedAndRemoved(t *testing.T) {
	// current playlist has 1 and 2
	mockRepo := &mockPlaylistRepo{getPlaylistTracksResponse: []zspotify.PlaylistItem{makePI("1"), makePI("2")}}
	// desired rows are 2, 3, 4 (1 dropped out, 3 and 4 added)
	mockTrack := &mockTrackRepo{rows: []psongs.Song{{SpotifyId: "2"}, {SpotifyId: "3"}, {SpotifyId: "4"}}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populateThrowbackPlaylist(ctx, service, mockTrack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockRepo.replaceCalls != 1 {
		t.Fatalf("expected 1 replace call, got %d", mockRepo.replaceCalls)
	}

	expectedReplaced := []zspotify.ID{zspotify.ID("2"), zspotify.ID("3"), zspotify.ID("4")}
	if !reflect.DeepEqual(mockRepo.replaced, expectedReplaced) {
		t.Fatalf("unexpected replaced order: got %v, want %v", mockRepo.replaced, expectedReplaced)
	}

	expectedDescription := "The top tracks added to the dynamic playlist from Jan 02, 2023 to Apr 02, 2023!"
	if mockRepo.updatedDescription != expectedDescription {
		t.Fatalf("expected description %q, got %q", expectedDescription, mockRepo.updatedDescription)
	}
}

func TestPopulateThrowbackPlaylist_ScoreOrderReorder(t *testing.T) {
	// current playlist has 1, 2, 3
	mockRepo := &mockPlaylistRepo{getPlaylistTracksResponse: []zspotify.PlaylistItem{makePI("1"), makePI("2"), makePI("3")}}
	// desired rows have same set, but new score ranking: 3, 1, 2
	mockTrack := &mockTrackRepo{rows: []psongs.Song{{SpotifyId: "3"}, {SpotifyId: "1"}, {SpotifyId: "2"}}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populateThrowbackPlaylist(ctx, service, mockTrack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockRepo.replaceCalls != 1 {
		t.Fatalf("expected 1 replace call, got %d", mockRepo.replaceCalls)
	}

	expectedReplaced := []zspotify.ID{zspotify.ID("3"), zspotify.ID("1"), zspotify.ID("2")}
	if !reflect.DeepEqual(mockRepo.replaced, expectedReplaced) {
		t.Fatalf("unexpected replaced order: got %v, want %v", mockRepo.replaced, expectedReplaced)
	}

	expectedDescription := "The top tracks added to the dynamic playlist from Jan 02, 2023 to Apr 02, 2023!"
	if mockRepo.updatedDescription != expectedDescription {
		t.Fatalf("expected description %q, got %q", expectedDescription, mockRepo.updatedDescription)
	}
}

func TestPopulateThrowbackPlaylist_NoTracks(t *testing.T) {
	// current playlist has 1 and 2
	mockRepo := &mockPlaylistRepo{getPlaylistTracksResponse: []zspotify.PlaylistItem{makePI("1"), makePI("2")}}
	// empty desired rows (e.g. threshold not met)
	mockTrack := &mockTrackRepo{rows: []psongs.Song{}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populateThrowbackPlaylist(ctx, service, mockTrack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockRepo.replaceCalls != 0 {
		t.Fatalf("expected 0 replace calls, got %d", mockRepo.replaceCalls)
	}

	if len(mockRepo.replaced) != 0 {
		t.Fatalf("expected no replaced tracks, got: %v", mockRepo.replaced)
	}

	if mockRepo.updatedDescription != "" {
		t.Fatalf("expected no description update, got: %q", mockRepo.updatedDescription)
	}
}

func TestPopulateThrowbackPlaylist_NoChanges(t *testing.T) {
	// current playlist has 1, 2, 3
	mockRepo := &mockPlaylistRepo{getPlaylistTracksResponse: []zspotify.PlaylistItem{makePI("1"), makePI("2"), makePI("3")}}
	// desired rows are 1, 2, 3 -> tracks already in exact score order
	mockTrack := &mockTrackRepo{rows: []psongs.Song{{SpotifyId: "1"}, {SpotifyId: "2"}, {SpotifyId: "3"}}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populateThrowbackPlaylist(ctx, service, mockTrack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockRepo.replaceCalls != 0 {
		t.Fatalf("expected 0 replace calls when already in order, got %d", mockRepo.replaceCalls)
	}

	if len(mockRepo.replaced) != 0 {
		t.Fatalf("expected no replaced tracks, got: %v", mockRepo.replaced)
	}

	expectedDescription := "The top tracks added to the dynamic playlist from Jan 02, 2023 to Apr 02, 2023!"
	if mockRepo.updatedDescription != expectedDescription {
		t.Fatalf("expected description %q, got %q", expectedDescription, mockRepo.updatedDescription)
	}
}
