package cron

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
)

const varietyTrackLimit = 100

func RunPopulateVarietyPlaylist(ctx context.Context) error {
	spClient, err := spotify.NewSpotifyClient(ctx)
	if err != nil {
		return err
	}
	playlistService := domain.NewSpotifyPlaylistService(&sp.SpotifyPlaylistRepo{
		Client:   spClient,
		Playlist: domain.VarietyPlaylist,
	})

	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		return err
	}

	trackService := domain.NewTrackService(persistence.NewPostgresTrackRepository(pgConn))
	return populateVarietyPlaylist(ctx, playlistService, trackService.Repo)
}

func populateVarietyPlaylist(ctx context.Context, playlistService *domain.SpotifyPlaylistService, trackRepo domain.AddTrackRepository) error {
	tracksToAdd, err := trackRepo.GetRandomTracks(varietyTrackLimit)
	if err != nil {
		return err
	}

	if len(tracksToAdd) == 0 {
		log.Info("No tracks available; skipping variety playlist update")
		return nil
	}

	playlistItems, err := playlistService.Repo.GetPlaylistTracks(ctx)
	if err != nil {
		return err
	}

	currentSet := playlistItemsToSet(playlistItems)
	desiredOrder, desiredSet := songsToIDsAndSet(tracksToAdd)

	toRemove := diffToRemove(currentSet, desiredSet)
	toAdd := diffToAdd(currentSet, desiredOrder)

	if len(toRemove) > 0 {
		log.Infof("Removing %d obsolete tracks from variety playlist", len(toRemove))
		if err := playlistService.Repo.RemoveTracksFromPlaylist(ctx, toRemove); err != nil {
			log.Errorf("Failed to remove tracks from variety playlist: %v", err)
		}
	}

	if len(toAdd) > 0 {
		log.Infof("Adding %d new tracks to variety playlist", len(toAdd))
		if err := playlistService.Repo.AddTracksToPlaylist(ctx, toAdd); err != nil {
			log.Errorf("Failed to add tracks to variety playlist: %v", err)
			return err
		}
	}

	log.Infof("Finished populating variety playlist. Added: %d, Removed: %d, Total desired: %d", len(toAdd), len(toRemove), len(desiredOrder))
	return nil
}
