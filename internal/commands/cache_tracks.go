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
	trackChan := make(chan *spotify.FullTrack)
	go func(c chan<- error) {
		err := commands.GetPlaylistTracks(ctx, trackChan)
		if err != nil {
			c <- err
		}
	}(errorChan)

	err := <-errorChan
	close(errorChan)
	if err != nil {
		return err
	}

	var fullTracks []*spotify.FullTrack
	for track := range trackChan {
		fullTracks = append(fullTracks, track)
	}
	close(trackChan)

	tracks := make([]types.CacheTrack, len(fullTracks))
	for i, track := range fullTracks {
		tracks[i] = types.CacheTrack{
			TrackId: track.ID.String(),
			AddedAt: time.Now().UTC(),
		}
	}

	err = re.Flush(ctx)
	if err != nil {
		return err
	}

	trackMap := map[string]string{}
	for _, track := range tracks {
		trackMap[track.TrackId] = track.AddedAt.Format(time.RFC3339)
	}

	err = re.SetMulti(ctx, trackMap)
	if err != nil {
		return err
	}

	return nil
}
