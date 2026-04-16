package cron

import (
	"context"
	"testing"

	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	psongs "github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	zspotify "github.com/zmb3/spotify/v2"
)

func TestPopulateVarietyPlaylist(t *testing.T) {
	// current playlist has 1 and 2
	mockRepo := &mockPlaylistRepo{getPlaylistTracksResponse: []zspotify.PlaylistItem{makePI("1"), makePI("2")}}
	// desired rows are 2,3,4 -> should remove 1, add 3,4, keep 2
	mockTrack := &mockTrackRepo{rows: []psongs.Song{{SpotifyId: "2"}, {SpotifyId: "3"}, {SpotifyId: "4"}}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populateVarietyPlaylist(ctx, service, mockTrack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// verify removed contains 1
	if !containsID(mockRepo.removed, zspotify.ID("1")) || len(mockRepo.removed) != 1 {
		t.Fatalf("unexpected removed: %v", mockRepo.removed)
	}

	// verify added contains 3 and 4
	if !containsID(mockRepo.added, zspotify.ID("3")) || !containsID(mockRepo.added, zspotify.ID("4")) || len(mockRepo.added) != 2 {
		t.Fatalf("unexpected added: %v", mockRepo.added)
	}
}

func TestPopulateVarietyPlaylist_NoChanges(t *testing.T) {
	// current playlist has 1,2,3
	mockRepo := &mockPlaylistRepo{getPlaylistTracksResponse: []zspotify.PlaylistItem{makePI("1"), makePI("2"), makePI("3")}}
	// desired rows are 1,2,3 -> should perform no adds/removes
	mockTrack := &mockTrackRepo{rows: []psongs.Song{{SpotifyId: "1"}, {SpotifyId: "2"}, {SpotifyId: "3"}}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populateVarietyPlaylist(ctx, service, mockTrack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mockRepo.removed) != 0 {
		t.Fatalf("expected no removed tracks, got: %v", mockRepo.removed)
	}

	if len(mockRepo.added) != 0 {
		t.Fatalf("expected no added tracks, got: %v", mockRepo.added)
	}
}

func TestPopulateVarietyPlaylist_Empty(t *testing.T) {
	mockRepo := &mockPlaylistRepo{}
	mockTrack := &mockTrackRepo{rows: []psongs.Song{}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populateVarietyPlaylist(ctx, service, mockTrack)
	if err != nil {
		t.Fatalf("expected no error on empty tracks, got %v", err)
	}
}
