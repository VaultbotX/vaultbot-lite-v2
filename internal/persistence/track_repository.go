package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/archive"
	artists2 "github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/artists"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/genres"
	psongs "github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/songs"
	"github.com/zmb3/spotify/v2"
)

type PostgresTrackRepository struct {
	db *sqlx.DB
}

func NewPostgresTrackRepository(db *sqlx.DB) *PostgresTrackRepository {
	return &PostgresTrackRepository{
		db: db,
	}
}

func (r *PostgresTrackRepository) AddTrackToDatabase(track *spotify.FullTrack, artists []*spotify.FullArtist) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	var allGenreIds []int
	var allArtistIds []int
	for _, artist := range artists {
		var genreIds []int
		for _, genre := range artist.Genres {
			addGenre, err := genres.AddGenre(tx, genre)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			genreIds = append(genreIds, addGenre.Id)
		}

		addArtist, err := artists2.AddArtist(tx, artist.ID.String(), artist.Name, genreIds)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		allGenreIds = append(allGenreIds, genreIds...)
		allArtistIds = append(allArtistIds, addArtist.Id)
	}

	addTrack, err := psongs.AddSong(tx, track, allGenreIds, allArtistIds)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = archive.AddArchive(tx, addTrack.Id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *PostgresTrackRepository) GetRandomGenreTracks() (songs []psongs.Song, genreName string, err error) {
	genre, err := genres.GetRandomGenre(r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", nil
		}
		return nil, "", err
	}

	tracks, err := psongs.GetTopSongsByGenre(r.db, genre.Id)
	if err != nil {
		return nil, "", err
	}

	return tracks, genre.Name, nil
}

func (r *PostgresTrackRepository) GetTop50Tracks() (songs []psongs.Song, err error) {
	return psongs.GetOverallTopSongs(r.db, 50)
}

// HasRecentArchiveEntry returns true if a song_archive row exists for the given
// Spotify track ID with created_at >= since. Used by the poll job to detect
// whether a playlist item has already been recorded for this addition event.
func (r *PostgresTrackRepository) HasRecentArchiveEntry(ctx context.Context, spotifyId string, since time.Time) (bool, error) {
	var exists bool
	err := r.db.QueryRowxContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM song_archive sa
			JOIN songs s ON sa.song_id = s.id
			WHERE s.spotify_id = $1
			AND sa.created_at >= $2
		)
	`, spotifyId, since).Scan(&exists)
	return exists, err
}
