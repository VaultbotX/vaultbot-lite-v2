package tracks

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/zmb3/spotify/v2"
	"time"
)

func PurgeTracks(ctx context.Context, preferenceService *domain.PreferenceService, spotifyPlaylistService *domain.SpotifyPlaylistService) error {
	tracks := persistence.TrackCache.GetAll()
	pref, err := preferenceService.Repo.Get(ctx, domain.MaxTrackAgeKey)
	if err != nil {
		return err
	}

	num, err := pref.IntValue()
	if err != nil {
		return err
	}

	maxTrackAge := time.Duration(num) * time.Millisecond
	oldestAllowed := time.Now().UTC().Add(-maxTrackAge)
	log.Debugf("Threshold: %s", oldestAllowed)
	var expiredTracks []spotify.ID
	for _, track := range tracks {
		if track.AddedAt.Before(oldestAllowed) {
			expiredTracks = append(expiredTracks, track.TrackId)
		}
	}

	if len(expiredTracks) == 0 {
		log.Debug("No expired tracks found")
		return nil
	}

	log.Debugf("Found %d expired tracks", len(expiredTracks))

	err = RemoveTracks(ctx, expiredTracks, spotifyPlaylistService)
	if err != nil {
		return err
	}

	return nil
}
