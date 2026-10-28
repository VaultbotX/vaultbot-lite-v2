package cron

import (
	"context"
	"slices"

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

	currentOrder := playlistItemsToIDs(playlistItems)
	desiredOrder, _ := songsToIDsAndSet(tracksToAdd)

	if slices.Equal(currentOrder, desiredOrder) {
		log.Info("High scores playlist is already in exact score order; no update needed")
		return nil
	}

	log.Infof("Updating high scores playlist with %d tracks in score order", len(desiredOrder))
	if err := playlistService.Repo.ReplacePlaylistTracks(ctx, desiredOrder); err != nil {
		log.Errorf("Failed to replace tracks in high scores playlist: %v", err)
		return err
	}

	log.Infof("Finished populating high scores playlist. Total tracks: %d", len(desiredOrder))
	return nil
}
