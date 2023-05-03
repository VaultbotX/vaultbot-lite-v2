package commands

import (
	"context"
	log "github.com/sirupsen/logrus"
	re "github.com/vaultbotx/vaultbot-lite/internal/database/redis"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	"github.com/zmb3/spotify/v2"
)

func RemoveTracks(ctx context.Context, trackIds []string) error {
	spotifyIds := make([]spotify.ID, len(trackIds))
	for i, trackId := range trackIds {
		spotifyIds[i] = spotify.ID(trackId)
	}
	log.Debug("Removing ", len(spotifyIds), " tracks from playlist")
	err := sp.RemoveTracksFromPlaylist(ctx, spotifyIds)
	if err != nil {
		return err
	}

	log.Debug("Removing ", len(trackIds), " tracks from cache")
	err = re.RemoveMulti(ctx, trackIds)
	if err != nil {
		return err
	}

	return nil
}
