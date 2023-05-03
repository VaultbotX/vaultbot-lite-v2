package commands

import (
	"context"
	re "github.com/vaultbotx/vaultbot-lite/internal/database/redis"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	"github.com/zmb3/spotify/v2"
)

func RemoveTracks(ctx context.Context, trackIds []string) error {
	spotifyIds := make([]spotify.ID, len(trackIds))
	for i, trackId := range trackIds {
		spotifyIds[i] = spotify.ID(trackId)
	}
	err := sp.RemoveTracksFromPlaylist(ctx, spotifyIds)
	if err != nil {
		return err
	}

	err = re.RemoveMulti(ctx, trackIds)
	if err != nil {
		return err
	}

	return nil
}
