package cron

import (
	psongs "github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	zspotify "github.com/zmb3/spotify/v2"
)

// diffToRemove returns IDs present in currentSet but not in desiredSet
func diffToRemove(currentSet, desiredSet map[zspotify.ID]struct{}) []zspotify.ID {
	var out []zspotify.ID
	for id := range currentSet {
		if _, ok := desiredSet[id]; !ok {
			out = append(out, id)
		}
	}
	return out
}

// diffToAdd returns IDs present in desiredOrder (in order) but not in currentSet
func diffToAdd(currentSet map[zspotify.ID]struct{}, desiredOrder []zspotify.ID) []zspotify.ID {
	var out []zspotify.ID
	for _, id := range desiredOrder {
		if _, ok := currentSet[id]; !ok {
			out = append(out, id)
		}
	}
	return out
}

// playlistItemsToSet converts a slice of spotify.PlaylistItem into a set of spotify IDs.
// It skips items that don't contain a track or have empty IDs.
func playlistItemsToSet(items []zspotify.PlaylistItem) map[zspotify.ID]struct{} {
	set := make(map[zspotify.ID]struct{})
	for _, item := range items {
		if item.Track.Track != nil && item.Track.Track.ID != "" {
			set[item.Track.Track.ID] = struct{}{}
		}
	}
	return set
}

// songsToIDsAndSet converts a slice of songs (as returned by the repository) into
// an ordered slice of spotify IDs and a set for quick membership checks.
func songsToIDsAndSet(rows []psongs.Song) ([]zspotify.ID, map[zspotify.ID]struct{}) {
	order := make([]zspotify.ID, 0, len(rows))
	set := make(map[zspotify.ID]struct{})
	for _, r := range rows {
		if r.SpotifyId == "" {
			continue
		}
		id := zspotify.ID(r.SpotifyId)
		set[id] = struct{}{}
		order = append(order, id)
	}
	return order, set
}
