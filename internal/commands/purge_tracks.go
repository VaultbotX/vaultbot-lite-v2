package commands

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/database"
	"github.com/zmb3/spotify/v2"
	"time"
)

func PurgeTracks(ctx context.Context) error {
	log.Debug("Purging tracks")

	tracks := database.Cache.GetAll()

	var expiredTracks []spotify.ID
	now := time.Now().UTC()
	for _, track := range tracks {
		// TODO: Make this configurable
		if track.AddedAt.Before(now.AddDate(0, 0, -14)) {
			expiredTracks = append(expiredTracks, track.TrackId)
		}
	}
	log.Debug("Found ", len(expiredTracks), " expired tracks")

	err := RemoveTracks(ctx, expiredTracks)
	if err != nil {
		return err
	}

	return nil
}
