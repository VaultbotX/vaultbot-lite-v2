package tracks

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	sp "github.com/zmb3/spotify/v2"
)

func CacheTracks(ctx context.Context, playlistService *domain.SpotifyPlaylistService) error {
	log.Debug("Caching tracks")
	playlistItems, err := playlistService.Repo.GetPlaylistTracks(ctx)
	if err != nil {
		return err
	}

	tracks := make([]*domain.CacheTrack, len(playlistItems))
	for i, track := range playlistItems {
		addedAt, err := time.Parse(sp.TimestampLayout, track.AddedAt)
		if err != nil {
			return err
		}

		tracks[i] = &domain.CacheTrack{
			TrackId: track.Track.Track.ID,
			AddedAt: addedAt,
		}
	}
	log.Debug("Found ", len(tracks), " tracks in playlist")
	log.Debug("Setting cache with new tracks")
	persistence.TrackCache.SetMulti(tracks)

	return nil
}
