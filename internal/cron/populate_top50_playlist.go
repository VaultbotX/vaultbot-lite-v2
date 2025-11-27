package cron

import (
	"context"
	"time"

	"github.com/go-co-op/gocron/v2"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
)

var (
	populateTop50PlaylistJob gocron.Job
)

func PopulateTop50Playlist(scheduler gocron.Scheduler) {
	job, err := scheduler.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(0, 0, 0),
			),
		),
		gocron.NewTask(populateTop50PlaylistOuter),
	)

	if err != nil {
		log.Fatalf("Failed to schedule populate genre playlist job: %v", err)
	}

	populateTop50PlaylistJob = job
}

func populateTop50PlaylistOuter() {
	newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	spClient, err := spotify.NewSpotifyClient(newCtx)
	if err != nil {
		log.Error(err)
		return
	}
	playlistService := domain.NewSpotifyPlaylistService(&sp.SpotifyPlaylistRepo{
		Client:   spClient,
		Playlist: domain.Top50Playlist,
	})

	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		log.Error(err)
		return
	}

	trackService := domain.NewTrackService(persistence.NewPostgresTrackRepository(pgConn))

	if err := populateTop50Playlist(newCtx, playlistService, trackService.Repo); err != nil {
		log.Errorf("populateTop50PlaylistOuter failed: %v", err)
	}
}

func populateTop50Playlist(ctx context.Context, playlistService *domain.SpotifyPlaylistService, trackRepo domain.AddTrackRepository) error {
	playlistItems, err := playlistService.Repo.GetPlaylistTracks(ctx)
	if err != nil {
		return err
	}

	tracksToAdd, err := trackRepo.GetTop50Tracks()
	if err != nil {
		return err
	}

	currentSet := playlistItemsToSet(playlistItems)
	desiredOrder, desiredSet := songsToIDsAndSet(tracksToAdd)

	toRemove := diffToRemove(currentSet, desiredSet)
	toAdd := diffToAdd(currentSet, desiredOrder)

	// Remove obsolete tracks
	if len(toRemove) > 0 {
		log.Infof("Removing %d obsolete tracks from top 50 playlist", len(toRemove))
		if err := playlistService.Repo.RemoveTracksFromPlaylist(ctx, toRemove); err != nil {
			log.Errorf("Failed to remove tracks from playlist: %v", err)
			// continue to attempt adding new tracks
		}
	}

	// Add new tracks
	if len(toAdd) > 0 {
		log.Infof("Adding %d new tracks to top 50 playlist", len(toAdd))
		if err := playlistService.Repo.AddTracksToPlaylist(ctx, toAdd); err != nil {
			log.Errorf("Failed to add tracks to playlist: %v", err)
			return err
		}
	}

	log.Infof("Finished populating top 50 playlist. Added: %d, Removed: %d, Total desired: %d", len(toAdd), len(toRemove), len(desiredOrder))
	return nil
}
