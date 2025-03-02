package persistence

import (
	"github.com/jmoiron/sqlx"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/users"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"github.com/zmb3/spotify/v2"
	"time"
)

type PostgresTrackRepository struct {
	db *sqlx.DB
}

func NewPostgresTrackRepository(db *sqlx.DB) *PostgresTrackRepository {
	return &PostgresTrackRepository{
		db: db,
	}
}

func (r *PostgresTrackRepository) AddTrackToDatabase(fields *types.UserFields, track *spotify.FullTrack, artist []*spotify.FullArtist, audioFeatures *spotify.AudioFeatures) error {
	tx, err := r.db.Beginx()
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

	// for each genre associated with song, album - insert new records + a link record
	// for each artist associated with song, album - insert new records + a link record
	// insert song if it doesn't exist, add links
	// always add to archive table

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
