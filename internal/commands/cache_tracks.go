package commands

import (
	"context"
	re "github.com/vaultbotx/vaultbot-lite/internal/database/redis"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
	"time"
)

func CacheTracks(ctx context.Context) error {
	errorChan := make(chan error)
	playlistItemChan := make(chan *spotify.PlaylistItem)
	go func(c chan<- error) {
		err := commands.GetPlaylistTracks(ctx, playlistItemChan)
		if err != nil {
			c <- err
		}
	}(errorChan)

	err := <-errorChan
	close(errorChan)
	if err != nil {
		return err
	}

	var playlistItems []*spotify.PlaylistItem
	for item := range playlistItemChan {
		playlistItems = append(playlistItems, item)
	}
	close(playlistItemChan)

	tracks := make([]*types.CacheTrack, len(playlistItems))
	for i, track := range playlistItems {
		var addedAt time.Time
		addedAt, err = time.Parse(spotify.TimestampLayout, track.AddedAt)
		if err != nil {
			return err
		}

		tracks[i] = &types.CacheTrack{
			TrackId: track.Track.Track.ID.String(),
			AddedAt: addedAt,
		}
	}

	err = re.Flush(ctx)
	if err != nil {
		return err
	}

	err = re.SetMulti(ctx, tracks)
	if err != nil {
		return err
	}

	return nil
}
