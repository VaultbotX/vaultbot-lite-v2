package tracks

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	"github.com/zmb3/spotify/v2"
)

func RemoveTracks(ctx context.Context, trackIds []spotify.ID) error {
	log.Debug("Removing ", len(trackIds), " tracks from playlist")
	err := sp.RemoveTracksFromPlaylist(ctx, trackIds)
	if err != nil {
		return err
	}

	log.Debug("Removing ", len(trackIds), " tracks from cache")
	persistence.TrackCache.RemoveMulti(trackIds)

	return nil
}
