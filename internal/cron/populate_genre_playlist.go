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
	populateGenrePlaylistJob gocron.Job
)

func RunPopulateGenrePlaylist() error {
	err := populateGenrePlaylistJob.RunNow()
	if err != nil {
		return err
	}

	return nil
}

func PopulateGenrePlaylist(scheduler gocron.Scheduler) {
	job, err := scheduler.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(0, 0, 0),
			),
		),
		gocron.NewTask(populatePlaylistOuter),
	)

	if err != nil {
		log.Fatalf("Failed to schedule populate genre playlist job: %v", err)
	}

	populateGenrePlaylistJob = job
}

func populatePlaylistOuter() {
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

	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		log.Error(err)
		return
	}

	trackService := domain.NewTrackService(persistence.NewPostgresTrackRepository(pgConn))

	if err := populatePlaylist(newCtx, playlistService, trackService.Repo); err != nil {
		log.Errorf("populatePlaylistOuter failed: %v", err)
	}
}

var (
	baseGenrePlaylistDescription = "A randomly selected genre tracked by Vaultbot. Revived as of 11/25/25 :)"
)

func populatePlaylist(ctx context.Context, playlistService *domain.SpotifyPlaylistService, trackRepo domain.AddTrackRepository) error {
	playlistItems, err := playlistService.Repo.GetPlaylistTracks(ctx)
	if err != nil {
		return err
	}

	tracksToAdd, genreName, err := trackRepo.GetRandomGenreTracks()
	if err != nil {
		return err
	}

	currentSet := playlistItemsToSet(playlistItems)
	desiredOrder, desiredSet := songsToIDsAndSet(tracksToAdd)

	toRemove := diffToRemove(currentSet, desiredSet)
	toAdd := diffToAdd(currentSet, desiredOrder)

	// Remove obsolete tracks
	if len(toRemove) > 0 {
		log.Infof("Removing %d obsolete tracks from genre playlist", len(toRemove))
		if err := playlistService.Repo.RemoveTracksFromPlaylist(ctx, toRemove); err != nil {
			log.Errorf("Failed to remove tracks from playlist: %v", err)
			// continue to attempt adding new tracks
		}
	}

	// Add new tracks
	if len(toAdd) > 0 {
		log.Infof("Adding %d new tracks to genre playlist", len(toAdd))
		if err := playlistService.Repo.AddTracksToPlaylist(ctx, toAdd); err != nil {
			log.Errorf("Failed to add tracks to playlist: %v", err)
			return err
		}
	}

	// Update playlist description
	newDescription := baseGenrePlaylistDescription
	if genreName != "" {
		newDescription += " Current genre: " + genreName + "."
	}

	if err := playlistService.Repo.UpdatePlaylistDescription(ctx, newDescription); err != nil {
		log.Errorf("Failed to update playlist description: %v", err)
		// Not a critical error, continue
	}

	log.Infof("Finished populating genre playlist. Added: %d, Removed: %d, Total desired: %d", len(toAdd), len(toRemove), len(desiredOrder))
	return nil
}
