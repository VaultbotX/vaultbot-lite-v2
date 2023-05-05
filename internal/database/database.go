package database

import (
	"context"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
	"time"
)

func AddTrackToDatabase(ctx context.Context, track *spotify.FullTrack, artist []*spotify.FullArtist,
	audioFeatures []*spotify.AudioFeatures) error {
	// 1. Add to Cache
	now := time.Now().UTC()

	Cache.Set(&types.CacheTrack{
		TrackId: track.ID,
		AddedAt: now,
	})

	// TODO
	// 2. Add to Neo4j

	// 3. Add to Mongo
	return nil
}
