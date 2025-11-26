package cron

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
	psongs "github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	zspotify "github.com/zmb3/spotify/v2"
)

func PopulateGenrePlaylist(scheduler *gocron.Scheduler) {
	_, err := scheduler.Every(1).Day().At("00:00").Do(populatePlaylist)
	if err != nil {
		log.Fatalf("Failed to schedule populate genre playlist job: %v", err)
	}
}

func populatePlaylist() {
	newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	spClient, err := spotify.NewSpotifyClient(newCtx)
	if err != nil {
		log.Error(err)
		return
	}
	playlistService := domain.NewSpotifyPlaylistService(&sp.SpotifyPlaylistRepo{
		Client:   spClient,
		Playlist: domain.GenrePlaylist,
	})

	playlistItems, err := playlistService.Repo.GetPlaylistTracks(newCtx)
	if err != nil {
		log.Error(err)
		return
	}

	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		log.Error(err)
		return
	}

	trackService := domain.NewTrackService(persistence.NewPostgresTrackRepository(pgConn))
	tracksToAdd, err := trackService.Repo.GetRandomGenreTracks()
	if err != nil {
		log.Error(err)
		return
	}

	// Build sets of current and desired track IDs using helper functions
	currentSet := playlistItemsToSet(playlistItems)
	desiredOrder, desiredSet := songsToIDsAndSet(tracksToAdd)

	// Calculate tracks to remove and to add using helpers
	toRemove := diffToRemove(currentSet, desiredSet)
	toAdd := diffToAdd(currentSet, desiredOrder)

	// Remove obsolete tracks
	if len(toRemove) > 0 {
		log.Infof("Removing %d obsolete tracks from genre playlist", len(toRemove))
		if err := playlistService.Repo.RemoveTracksFromPlaylist(newCtx, toRemove); err != nil {
			log.Errorf("Failed to remove tracks from playlist: %v", err)
			// continue to attempt adding new tracks
		}
	}

	// Add new tracks
	if len(toAdd) > 0 {
		log.Infof("Adding %d new tracks to genre playlist", len(toAdd))
		if err := playlistService.Repo.AddTracksToPlaylist(newCtx, toAdd); err != nil {
			log.Errorf("Failed to add tracks to playlist: %v", err)
			return
		}
	}

	log.Infof("Finished populating genre playlist. Added: %d, Removed: %d, Total desired: %d", len(toAdd), len(toRemove), len(desiredOrder))
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
