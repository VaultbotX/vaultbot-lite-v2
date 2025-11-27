package cron

import (
	"testing"
	"time"

	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	"github.com/zmb3/spotify/v2"
)

func makePlaylistItem(id string) spotify.PlaylistItem {
	return spotify.PlaylistItem{
		AddedAt: time.Now().Format(spotify.TimestampLayout),
		Track: spotify.PlaylistItemTrack{
			Track: &spotify.FullTrack{
				SimpleTrack: spotify.SimpleTrack{ID: spotify.ID(id)},
			},
		},
	}
}

// Test playlistItemsToSet with typical items and an item missing a track
func TestPlaylistItemsToSet(t *testing.T) {
	items := []spotify.PlaylistItem{
		makePlaylistItem("1"),
		makePlaylistItem("2"),
		// item with nil track
		{Track: spotify.PlaylistItemTrack{Track: nil}},
	}

	set := playlistItemsToSet(items)
	if len(set) != 2 {
		t.Fatalf("expected set length 2, got %d", len(set))
	}
	if _, ok := set[spotify.ID("1")]; !ok {
		t.Fatalf("expected id 1 in set")
	}
	if _, ok := set[spotify.ID("2")]; !ok {
		t.Fatalf("expected id 2 in set")
	}
}

// Test songsToIDsAndSet handles empty SpotifyId and preserves order
func TestSongsToIDsAndSet(t *testing.T) {
	rows := []songs.Song{{SpotifyId: "a"}, {SpotifyId: ""}, {SpotifyId: "b"}, {SpotifyId: "c"}}

	order, set := songsToIDsAndSet(rows)
	if len(order) != 3 {
		t.Fatalf("expected order len 3, got %d", len(order))
	}
	if order[0] != spotify.ID("a") || order[1] != spotify.ID("b") || order[2] != spotify.ID("c") {
		t.Fatalf("order mismatch: %+v", order)
	}
	if len(set) != 3 {
		t.Fatalf("expected set len 3, got %d", len(set))
	}
}

// Test songsToIDsAndSet handles an empty slice (should return empty order and set)
func TestSongsToIDsAndSet_EmptyRows(t *testing.T) {
	var rows []songs.Song

	order, set := songsToIDsAndSet(rows)
	if len(order) != 0 {
		t.Fatalf("expected empty order for empty rows, got %d", len(order))
	}
	if len(set) != 0 {
		t.Fatalf("expected empty set for empty rows, got %d", len(set))
	}
}

// Test diffToRemove identifies items present in current but not desired
func TestDiffToRemove(t *testing.T) {
	current := map[spotify.ID]struct{}{spotify.ID("1"): {}, spotify.ID("2"): {}, spotify.ID("3"): {}}
	desired := map[spotify.ID]struct{}{spotify.ID("2"): {}, spotify.ID("4"): {}}

	rem := diffToRemove(current, desired)
	// rem should contain 1 and 3 in any order
	if len(rem) != 2 {
		t.Fatalf("expected 2 to remove, got %d", len(rem))
	}
	m := map[spotify.ID]struct{}{}
	for _, id := range rem {
		m[id] = struct{}{}
	}
	if _, ok := m[spotify.ID("1")]; !ok {
		t.Fatalf("expected 1 in remove list")
	}
	if _, ok := m[spotify.ID("3")]; !ok {
		t.Fatalf("expected 3 in remove list")
	}
}

// Test diffToAdd returns desired order items not present in current
func TestDiffToAdd(t *testing.T) {
	current := map[spotify.ID]struct{}{spotify.ID("1"): {}, spotify.ID("3"): {}}
	desiredOrder := []spotify.ID{spotify.ID("2"), spotify.ID("3"), spotify.ID("4")}

	add := diffToAdd(current, desiredOrder)
	if len(add) != 2 {
		t.Fatalf("expected 2 to add, got %d", len(add))
	}
	if add[0] != spotify.ID("2") || add[1] != spotify.ID("4") {
		t.Fatalf("unexpected add order: %+v", add)
	}
}
