package database

import (
	"context"
	mongocommands "github.com/vaultbotx/vaultbot-lite/internal/database/mongo/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
	"time"
)

func AddTrackToDatabase(ctx context.Context, fields *types.UserFields, track *spotify.FullTrack,
	artist []*spotify.FullArtist, audioFeatures []*spotify.AudioFeatures) error {
	// 1. Add to TrackCache
	now := time.Now()

	// TODO: use redis here
	TrackCache.Set(&types.CacheTrack{
		TrackId: track.ID,
		AddedAt: now.UTC(),
	})

	// TODO: get rid of mongo dep and just use postgres
	// 2. Add to Mongo
	err := mongocommands.AddTrack(ctx, track.ID, fields, now)
	if err != nil {
		return err
	}

	return nil
}
