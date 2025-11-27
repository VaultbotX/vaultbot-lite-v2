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
	populateHighScoresPlaylistJob gocron.Job
)

func RunPopulateHighScoresPlaylist() error {
	err := populateHighScoresPlaylistJob.RunNow()
	if err != nil {
		return err
	}

	return nil
}

func PopulateHighScoresPlaylist(scheduler gocron.Scheduler) {
	job, err := scheduler.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(0, 0, 0),
			),
		),
		gocron.NewTask(populateHighScoresPlaylistOuter),
	)

	if err != nil {
		log.Fatalf("Failed to schedule populate high scores playlist job: %v", err)
	}

	populateHighScoresPlaylistJob = job
}

func populateHighScoresPlaylistOuter() {
	newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	spClient, err := spotify.NewSpotifyClient(newCtx)
	if err != nil {
		log.Error(err)
		return
	}
	playlistService := domain.NewSpotifyPlaylistService(&sp.SpotifyPlaylistRepo{
		Client:   spClient,
		Playlist: domain.HighScoresPlaylist,
	})

	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		log.Error(err)
		return
	}

	trackService := domain.NewTrackService(persistence.NewPostgresTrackRepository(pgConn))

	if err := populateHighScoresPlaylist(newCtx, playlistService, trackService.Repo); err != nil {
		log.Errorf("populateHighScoresPlaylistOuter failed: %v", err)
	}
}

func populateHighScoresPlaylist(ctx context.Context, playlistService *domain.SpotifyPlaylistService, trackRepo domain.AddTrackRepository) error {
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
		log.Infof("Removing %d obsolete tracks from high scores playlist", len(toRemove))
		if err := playlistService.Repo.RemoveTracksFromPlaylist(ctx, toRemove); err != nil {
			log.Errorf("Failed to remove tracks from playlist: %v", err)
			// continue to attempt adding new tracks
		}
	}

	// Add new tracks
	if len(toAdd) > 0 {
		log.Infof("Adding %d new tracks to high scores playlist", len(toAdd))
		if err := playlistService.Repo.AddTracksToPlaylist(ctx, toAdd); err != nil {
			log.Errorf("Failed to add tracks to playlist: %v", err)
			return err
		}
	}

	log.Infof("Finished populating high scores playlist. Added: %d, Removed: %d, Total desired: %d", len(toAdd), len(toRemove), len(desiredOrder))
	return nil
}
