package persistence

import (
	"github.com/jmoiron/sqlx"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/archive"
	artists2 "github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/artists"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/genres"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/users"
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

func (r *PostgresTrackRepository) AddTrackToDatabase(fields *domain.UserFields, track *spotify.FullTrack, artists []*spotify.FullArtist) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	addUser, err := users.AddUser(tx, fields)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			return err2
		}
		return err
	}

	var allGenreIds []int
	var allArtistIds []int
	for _, artist := range artists {
		var genreIds []int
		for _, genre := range artist.Genres {
			// insert genre if it doesn't exist
			addGenre, err := genres.AddGenre(tx, genre)
			if err != nil {
				return err
			}
			genreIds = append(genreIds, addGenre.Id)
		}

		// insert artist if it doesn't exist, add links
		addArtist, err := artists2.AddArtist(tx, artist.ID.String(), artist.Name, genreIds)
		if err != nil {
			return err
		}
		allGenreIds = append(allGenreIds, genreIds...)
		allArtistIds = append(allArtistIds, addArtist.Id)
	}

	addTrack, err := songs.AddSong(tx, track, allGenreIds, allArtistIds)
	if err != nil {
		return err
	}

	_, err = archive.AddArchive(tx, addTrack.Id, addUser.Id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	now := time.Now()
	TrackCache.Set(&domain.CacheTrack{
		TrackId: track.ID,
		AddedAt: now.UTC(),
	})

	return nil
}
