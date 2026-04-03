package cron

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/tracks"
)

func RunPurge(ctx context.Context) error {
	spClient, err := spotify.NewSpotifyClient(ctx)
	if err != nil {
		return err
	}
	spPlaylistService := domain.NewSpotifyPlaylistService(&sp.SpotifyPlaylistRepo{
		Client:   spClient,
		Playlist: domain.DynamicPlaylist,
	})

	now := time.Now().UTC()
	count, err := tracks.PurgeTracks(ctx, now, spPlaylistService)
	if err != nil {
		return err
	}
	log.Infof("Purged %d tracks", count)
	return nil
}
