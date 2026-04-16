package cron

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
)

const (
	throwbackMinTrackCount          = 20
	baseThrowbackPlaylistDescription = "The best tracks from a single release year, as voted by the Vaultbot community."
)

func RunPopulateThrowbackPlaylist(ctx context.Context) error {
	spClient, err := spotify.NewSpotifyClient(ctx)
	if err != nil {
		return err
	}
	playlistService := domain.NewSpotifyPlaylistService(&sp.SpotifyPlaylistRepo{
		Client:   spClient,
		Playlist: domain.ThrowbackPlaylist,
	})

	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		return err
	}

	trackService := domain.NewTrackService(persistence.NewPostgresTrackRepository(pgConn))
	return populateThrowbackPlaylist(ctx, playlistService, trackService.Repo)
}

func populateThrowbackPlaylist(ctx context.Context, playlistService *domain.SpotifyPlaylistService, trackRepo domain.AddTrackRepository) error {
	tracksToAdd, year, err := trackRepo.GetTopYearTracks(throwbackMinTrackCount)
	if err != nil {
		return err
	}

	if len(tracksToAdd) == 0 {
		log.Info("No release year meets the minimum track threshold; skipping throwback playlist update")
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
		log.Infof("Removing %d obsolete tracks from throwback playlist", len(toRemove))
		if err := playlistService.Repo.RemoveTracksFromPlaylist(ctx, toRemove); err != nil {
			log.Errorf("Failed to remove tracks from throwback playlist: %v", err)
		}
	}

	if len(toAdd) > 0 {
		log.Infof("Adding %d new tracks to throwback playlist", len(toAdd))
		if err := playlistService.Repo.AddTracksToPlaylist(ctx, toAdd); err != nil {
			log.Errorf("Failed to add tracks to throwback playlist: %v", err)
			return err
		}
	}

	newDescription := baseThrowbackPlaylistDescription
	if year > 0 {
		newDescription += fmt.Sprintf(" Current year: %d.", year)
	}

	if err := playlistService.Repo.UpdatePlaylistDescription(ctx, newDescription); err != nil {
		log.Errorf("Failed to update throwback playlist description: %v", err)
	}

	log.Infof("Finished populating throwback playlist. Added: %d, Removed: %d, Total desired: %d. Year: %d", len(toAdd), len(toRemove), len(desiredOrder), year)
	return nil
}
