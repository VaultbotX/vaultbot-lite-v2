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
	// 1. Add to Cache
	now := time.Now()

	Cache.Set(&types.CacheTrack{
		TrackId: track.ID,
		AddedAt: now.UTC(),
	})

	// TODO
	// 2. Add to Neo4j

	// 3. Add to Mongo
	err := mongocommands.AddTrack(ctx, track.ID, fields, now)
	if err != nil {
		return err
	}

	return nil
}
