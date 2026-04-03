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

func RunPopulateHighScoresPlaylist(ctx context.Context) error {
	spClient, err := spotify.NewSpotifyClient(ctx)
	if err != nil {
		return err
	}
	playlistService := domain.NewSpotifyPlaylistService(&sp.SpotifyPlaylistRepo{
		Client:   spClient,
		Playlist: domain.HighScoresPlaylist,
	})

	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		return err
	}

	trackService := domain.NewTrackService(persistence.NewPostgresTrackRepository(pgConn))
	return populateHighScoresPlaylist(ctx, playlistService, trackService.Repo)
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

	if len(toRemove) > 0 {
		log.Infof("Removing %d obsolete tracks from high scores playlist", len(toRemove))
		if err := playlistService.Repo.RemoveTracksFromPlaylist(ctx, toRemove); err != nil {
			log.Errorf("Failed to remove tracks from playlist: %v", err)
		}
	}

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
