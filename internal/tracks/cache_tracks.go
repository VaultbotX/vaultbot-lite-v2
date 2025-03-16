package tracks

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	sp "github.com/zmb3/spotify/v2"
	"time"
)

func CacheTracks(ctx context.Context, playlistService *domain.SpotifyPlaylistService) error {
	log.Debug("Caching tracks")
	errorChan := make(chan error)
	playlistItemChan := make(chan *sp.PlaylistItem)

	go func(c chan<- error) {
		err := playlistService.Repo.GetPlaylistTracks(playlistItemChan, ctx)
		if err != nil {
			c <- err
		}
	}(errorChan)

	var playlistItems []*sp.PlaylistItem
	done := false
	for !done {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errorChan:
			close(errorChan)
			if err != nil {
				return err
			}
			break
		case track, ok := <-playlistItemChan:
			if !ok {
				done = true
				break
			}
			playlistItems = append(playlistItems, track)
		}
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
