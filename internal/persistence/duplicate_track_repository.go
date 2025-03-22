package persistence

import (
	"github.com/jmoiron/sqlx"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/zmb3/spotify/v2"
)

type PostgresDuplicateTrackRepository struct {
	db *sqlx.DB
}

func NewPostgresDuplicateTrackRepository(db *sqlx.DB) *PostgresDuplicateTrackRepository {
	return &PostgresDuplicateTrackRepository{
		db: db,
	}
}

type DuplicateSongMapping struct {
	SourceSongSpotifyId string `db:"source_song_spotify_id"`
	TargetSongSpotifyId string `db:"target_song_spotify_id"`
}

// TODO: model for the return type of the first query or just parse the row out
func (r *PostgresDuplicateTrackRepository) GetRelatedTracks(trackId spotify.ID) ([]domain.TrackPartialWithMetadata, error) {
	panic("not implemented")
	//var relatedTracks []domain.TrackPartialWithMetadata
	//
	//query := `
	//    SELECT s.id,
	//           s.name,
	//           s.duration, -- may not have this modeled on the db yet for some reason...
	//           s.tempo, -- may not have this modeled on the db yet for some reason...
	//           s.album,
	//           s.release_date
	//    FROM songs s
	//             JOIN link_song_artists "as" on s.id = "as".song_id
	//    WHERE "as".artist_id = ANY (SELECT as2.artist_id
	//                                FROM link_song_artists as2
	//                                JOIN songs s2 on as2.song_id = s2.id
	//                                WHERE s2.spotify_id = $1)
	//    GROUP BY s.id;
	//`

	//rows, err := r.db.Queryx(query, trackId.String())
	//if err != nil {
	//	return nil, err
	//}
	//defer rows.Close()
	//
	//for rows.Next() {
	//	var mapping DuplicateSongMapping
	//	if err := rows.StructScan(&mapping); err != nil {
	//		return nil, err
	//	}
	//	relatedTracks = append(relatedTracks, domain.TrackPartialWithMetadata{
	//		Id: 	  mapping.SourceSongSpotifyId,
	//		SpotifyId: spotify.ID(mapping.TargetSongSpotifyId)
	//	})
	//}

	//return relatedTracks, nil
}

func (r *PostgresDuplicateTrackRepository) SetDuplicateTrack(sourceTrackId spotify.ID, targetTrackId spotify.ID) error {
	panic("not implemented")
}
