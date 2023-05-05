package database

import (
	"context"
	re "github.com/vaultbotx/vaultbot-lite/internal/database/redis"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
	"time"
)

func AddTrackToDatabase(ctx context.Context, track *spotify.FullTrack, artist []*spotify.FullArtist,
	audioFeatures []*spotify.AudioFeatures) error {
	// 1. Add to redis cache
	now := time.Now().UTC()
	cacheTrack := types.CacheTrack{TrackId: track.ID.String(), AddedAt: now}
	err := re.Set(ctx, &cacheTrack)
	if err != nil {
		return err
	}

	// TODO
	// 2. Add to Neo4j

	// 3. Add to Mongo
	return nil
}
