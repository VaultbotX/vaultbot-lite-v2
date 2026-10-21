package cron

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
)

const (
	throwbackWindowDuration = 90 * 24 * time.Hour
	throwbackBufferDuration = 30 * 24 * time.Hour
	throwbackMinTrackCount  = 30
	throwbackTrackLimit     = 50
	throwbackMaxAttempts    = 3
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
	tracksToAdd, start, end, err := trackRepo.GetTopWindowTracks(throwbackWindowDuration, throwbackBufferDuration, throwbackMinTrackCount, throwbackTrackLimit, throwbackMaxAttempts)
	if err != nil {
		return err
	}

	if len(tracksToAdd) == 0 {
		log.Info("No window meets the minimum track threshold; skipping throwback playlist update")
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

	newDescription := fmt.Sprintf("The top tracks added to the dynamic playlist from %s to %s!",
		start.Format("Jan 02, 2006"),
		end.Format("Jan 02, 2006"),
	)

	if err := playlistService.Repo.UpdatePlaylistDescription(ctx, newDescription); err != nil {
		log.Errorf("Failed to update throwback playlist description: %v", err)
	}

	log.Infof("Finished populating throwback playlist. Added: %d, Removed: %d, Total desired: %d. Window: %s to %s",
		len(toAdd), len(toRemove), len(desiredOrder), start.Format("Jan 02, 2006"), end.Format("Jan 02, 2006"))
	return nil
}
