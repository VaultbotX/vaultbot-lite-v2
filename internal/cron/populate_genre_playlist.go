package cron

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
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

	// Build sets of current and desired track IDs
	currentSet := make(map[zspotify.ID]struct{})
	for _, item := range playlistItems {
		if item.Track.Track.ID != "" {
			currentSet[item.Track.Track.ID] = struct{}{}
		}
	}

	desiredSet := make(map[zspotify.ID]struct{})
	var desiredOrder []zspotify.ID
	for _, t := range tracksToAdd {
		if t.SpotifyId == "" {
			continue
		}
		id := zspotify.ID(t.SpotifyId)
		desiredSet[id] = struct{}{}
		desiredOrder = append(desiredOrder, id)
	}

	// Calculate tracks to remove (present but not desired)
	var toRemove []zspotify.ID
	for id := range currentSet {
		if _, ok := desiredSet[id]; !ok {
			toRemove = append(toRemove, id)
		}
	}

	// Calculate tracks to add (desired but not present)
	var toAdd []zspotify.ID
	for _, id := range desiredOrder {
		if _, ok := currentSet[id]; !ok {
			toAdd = append(toAdd, id)
		}
	}

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
