package persistence

import (
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/zmb3/spotify/v2"
	"sync"
	"time"
)

// TrackCache is a very simple in-memory cache for storing track IDs that are currently in the dynamic playlist
// Deprecated: in favor of Redis once that is implemented
var TrackCache = newCache()

type trackCache struct {
	data map[spotify.ID]*time.Time
	mu   sync.RWMutex
}

func newCache() *trackCache {
	return &trackCache{
		data: make(map[spotify.ID]*time.Time),
	}
}

func (c *trackCache) Get(key spotify.ID) *time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if val, ok := c.data[key]; ok {
		return val
	}

	return nil
}

func (c *trackCache) Set(track *domain.CacheTrack) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[track.TrackId] = &track.AddedAt
}

func (c *trackCache) SetMulti(tracks []*domain.CacheTrack) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, track := range tracks {
		c.data[track.TrackId] = &track.AddedAt
	}
}

func (c *trackCache) GetAll() []*domain.CacheTrack {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var tracks []*domain.CacheTrack
	for trackId, addedAt := range c.data {
		tracks = append(tracks, &domain.CacheTrack{
			TrackId: trackId,
			AddedAt: *addedAt,
		})
	}
	return tracks
}

func (c *trackCache) RemoveMulti(ids []spotify.ID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, id := range ids {
		delete(c.data, id)
	}
}
