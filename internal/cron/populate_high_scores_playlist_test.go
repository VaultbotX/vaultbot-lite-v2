package cron

import (
	"context"
	"reflect"
	"testing"

	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	psongs "github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	zspotify "github.com/zmb3/spotify/v2"
)

func TestPopulateHighScoresPlaylist_AddedAndRemoved(t *testing.T) {
	// current playlist has 1 and 2
	mockRepo := &mockPlaylistRepo{getPlaylistTracksResponse: []zspotify.PlaylistItem{makePI("1"), makePI("2")}}
	// desired rows are 2, 3, 4 (1 dropped out, 3 and 4 added)
	mockTrack := &mockTrackRepo{rows: []psongs.Song{{SpotifyId: "2"}, {SpotifyId: "3"}, {SpotifyId: "4"}}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populateHighScoresPlaylist(ctx, service, mockTrack)
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
}

func TestPopulateHighScoresPlaylist_ScoreOrderReorder(t *testing.T) {
	// current playlist has 1, 2, 3
	mockRepo := &mockPlaylistRepo{getPlaylistTracksResponse: []zspotify.PlaylistItem{makePI("1"), makePI("2"), makePI("3")}}
	// desired rows have same set, but new score ranking: 3, 1, 2
	mockTrack := &mockTrackRepo{rows: []psongs.Song{{SpotifyId: "3"}, {SpotifyId: "1"}, {SpotifyId: "2"}}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populateHighScoresPlaylist(ctx, service, mockTrack)
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
}

func TestPopulateHighScoresPlaylist_NoChanges(t *testing.T) {
	// current playlist has 1, 2, 3
	mockRepo := &mockPlaylistRepo{getPlaylistTracksResponse: []zspotify.PlaylistItem{makePI("1"), makePI("2"), makePI("3")}}
	// desired rows are already in exact score order 1, 2, 3
	mockTrack := &mockTrackRepo{rows: []psongs.Song{{SpotifyId: "1"}, {SpotifyId: "2"}, {SpotifyId: "3"}}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populateHighScoresPlaylist(ctx, service, mockTrack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockRepo.replaceCalls != 0 {
		t.Fatalf("expected 0 replace calls when already in order, got %d", mockRepo.replaceCalls)
	}

	if len(mockRepo.replaced) != 0 {
		t.Fatalf("expected no replaced tracks, got: %v", mockRepo.replaced)
	}
}
