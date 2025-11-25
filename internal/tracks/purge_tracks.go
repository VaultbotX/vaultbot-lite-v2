package tracks

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/zmb3/spotify/v2"
)

func PurgeTracks(ctx context.Context, nowUtc time.Time, preferenceService *domain.PreferenceService, spotifyPlaylistService *domain.SpotifyPlaylistService) (int, error) {
	tracks := persistence.TrackCache.GetAll()
	pref, err := preferenceService.Repo.Get(ctx, domain.MaxTrackAgeKey)
	if err != nil {
		return 0, err
	}

	num, err := pref.IntValue()
	if err != nil {
		return 0, err
	}

	maxTrackAge := time.Duration(num) * time.Millisecond
	oldestAllowed := nowUtc.Add(-maxTrackAge)
	log.Debugf("Threshold: %s", oldestAllowed)
	var expiredTracks []spotify.ID
	for _, track := range tracks {
		if track.AddedAt.Before(oldestAllowed) {
			expiredTracks = append(expiredTracks, track.TrackId)
		}
	}

	if len(expiredTracks) == 0 {
		log.Debug("No expired tracks found")
		return 0, nil
	}

	log.Debugf("Found %d expired tracks", len(expiredTracks))

	err = RemoveTracks(ctx, expiredTracks, spotifyPlaylistService)
	if err != nil {
		return 0, err
	}

	return len(expiredTracks), nil
}
