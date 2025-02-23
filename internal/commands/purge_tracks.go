package commands

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/zmb3/spotify/v2"
	"time"
)

func PurgeTracks(ctx context.Context) error {
	tracks := persistence.TrackCache.GetAll()
	pref, err := GetMaxTrackAgePreference()
	if err != nil {
		return err
	}

	maxTrackAge := time.Duration(pref.Value.(int32)) * time.Millisecond
	oldestAllowed := time.Now().UTC().Add(-maxTrackAge)
	log.Debugf("Threshold: %s", oldestAllowed)
	var expiredTracks []spotify.ID
	for _, track := range tracks {
		if track.AddedAt.Before(oldestAllowed) {
			expiredTracks = append(expiredTracks, track.TrackId)
		}
	}
	log.Debugf("Found %d expired tracks", len(expiredTracks))

	err = RemoveTracks(ctx, expiredTracks)
	if err != nil {
		return err
	}

	return nil
}
