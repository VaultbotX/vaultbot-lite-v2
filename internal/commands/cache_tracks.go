package commands

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
	"time"
)

func CacheTracks(ctx context.Context) error {
	log.Debug("Caching tracks")
	errorChan := make(chan error)
	playlistItemChan := make(chan *spotify.PlaylistItem)
	go func(c chan<- error) {
		err := commands.GetPlaylistTracks(ctx, playlistItemChan)
		if err != nil {
			c <- err
		}
	}(errorChan)

	var playlistItems []*spotify.PlaylistItem
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

	tracks := make([]*types.CacheTrack, len(playlistItems))
	for i, track := range playlistItems {
		addedAt, err := time.Parse(spotify.TimestampLayout, track.AddedAt)
		if err != nil {
			return err
		}

		tracks[i] = &types.CacheTrack{
			TrackId: track.Track.Track.ID,
			AddedAt: addedAt,
		}
	}
	log.Debug("Found ", len(tracks), " tracks in playlist")
	log.Debug("Setting cache with new tracks")
	persistence.TrackCache.SetMulti(tracks)

	return nil
}
