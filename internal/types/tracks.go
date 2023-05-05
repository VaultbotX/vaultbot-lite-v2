package types

import (
	"github.com/zmb3/spotify/v2"
	"time"
)

type CacheTrack struct {
	TrackId spotify.ID
	AddedAt time.Time
}
