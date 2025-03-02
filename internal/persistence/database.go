package persistence

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/users"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
	"time"
)

func AddTrackToDatabase(ctx context.Context, fields *types.UserFields, track *spotify.FullTrack,
	artist []*spotify.FullArtist, audioFeatures []*spotify.AudioFeatures) error {
	// TODO: some sort of dep inversion of this db
	db, err := sqlx.Connect("postgres", "TODO")
	if err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// TODO
	_, err = users.AddUser(tx, fields)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}

	// for each genre associated with song, album - insert

	err = tx.Commit()
	if err != nil {
		return err
	}

	now := time.Now()
	TrackCache.Set(&types.CacheTrack{
		TrackId: track.ID,
		AddedAt: now.UTC(),
	})

	return nil
}
