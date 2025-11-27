package cron

import (
	"context"
	"reflect"
	"testing"

	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	psongs "github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	zspotify "github.com/zmb3/spotify/v2"
)

func TestPopulateHighScoresPlaylist(t *testing.T) {
	// current playlist has 1 and 2
	mockRepo := &mockPlaylistRepo{getPlaylistTracksResponse: []zspotify.PlaylistItem{makePI("1"), makePI("2")}}
	// desired rows are 2,3,4 -> should remove 1, add 3,4, keep 2
	mockTrack := &mockTrackRepo{rows: []psongs.Song{{SpotifyId: "2"}, {SpotifyId: "3"}, {SpotifyId: "4"}}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populateHighScoresPlaylist(ctx, service, mockTrack)
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

func TestPopulateHighScoresPlaylist_NoChanges(t *testing.T) {
	// current playlist has 1,2,3
	mockRepo := &mockPlaylistRepo{getPlaylistTracksResponse: []zspotify.PlaylistItem{makePI("1"), makePI("2"), makePI("3")}}
	// desired rows are 1,2,3 -> should perform no adds/removes
	mockTrack := &mockTrackRepo{rows: []psongs.Song{{SpotifyId: "1"}, {SpotifyId: "2"}, {SpotifyId: "3"}}}

	service := &domain.SpotifyPlaylistService{Repo: mockRepo}

	ctx := context.Background()
	err := populateHighScoresPlaylist(ctx, service, mockTrack)
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
