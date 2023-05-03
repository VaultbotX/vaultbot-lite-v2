package commands

import (
	"context"
	log "github.com/sirupsen/logrus"
	re "github.com/vaultbotx/vaultbot-lite/internal/database/redis"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"time"
)

func PurgeTracks(ctx context.Context) error {
	log.Debug("Purging tracks")
	errChan := make(chan error)
	trackChan := make(chan *types.CacheTrack)

	go func(c chan<- error) {
		err := re.GetAll(ctx, trackChan)
		if err != nil {
			c <- err
		}

		close(c)
	}(errChan)

	err := <-errChan
	if err != nil {
		return err
	}

	var tracks []*types.CacheTrack
	for track := range trackChan {
		tracks = append(tracks, track)
	}
	close(trackChan)

	var expiredTracks []string
	now := time.Now().UTC()
	for _, track := range tracks {
		if track.AddedAt.Before(now.AddDate(0, 0, -14)) {
			expiredTracks = append(expiredTracks, track.TrackId)
		}
	}
	log.Debug("Found ", len(expiredTracks), " expired tracks")

	err = RemoveTracks(ctx, expiredTracks)
	if err != nil {
		return err
	}

	return nil
}
