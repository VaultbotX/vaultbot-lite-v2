package tracks

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/zmb3/spotify/v2"
)

const maxTrackAge = 14 * 24 * time.Hour

func PurgeTracks(ctx context.Context, nowUtc time.Time, spotifyPlaylistService *domain.SpotifyPlaylistService) (int, error) {
	playlistItems, err := spotifyPlaylistService.Repo.GetPlaylistTracks(ctx)
	if err != nil {
		return 0, err
	}

	oldestAllowed := nowUtc.Add(-maxTrackAge)
	log.Debugf("Purge threshold: %s", oldestAllowed)

	var expiredTracks []spotify.ID
	for _, item := range playlistItems {
		if item.Track.Track == nil || item.Track.Track.ID == "" {
			continue
		}
		addedAt, err := time.Parse(spotify.TimestampLayout, item.AddedAt)
		if err != nil {
			return 0, err
		}
		if addedAt.Before(oldestAllowed) {
			expiredTracks = append(expiredTracks, item.Track.Track.ID)
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
