package main

import (
	"context"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	spotify2 "github.com/zmb3/spotify/v2"
	"os"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	host, hostExists := os.LookupEnv("POSTGRES_HOST")
	port, portExists := os.LookupEnv("POSTGRES_PORT")
	user, userExists := os.LookupEnv("POSTGRES_USER")
	password, passwordExists := os.LookupEnv("POSTGRES_PASSWORD")

	if !hostExists || !portExists || !userExists || !passwordExists {
		log.Fatal("POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, and POSTGRES_PASSWORD must be set")
	}

	spotifyClient, err := spotify.NewSpotifyClient(context.Background())
	if err != nil {
		log.Fatalf("Error creating Spotify spotifyClient: %v", err)
	}

	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=vaultbot"
	_, envPresent := os.LookupEnv("ENVIRONMENT")
	if envPresent {
		// append sslmode=disable - local dev only
		dsn += " sslmode=disable"
	} else {
		// append sslmode=require - prod
		dsn += " sslmode=require"
	}

	db := sqlx.MustConnect("postgres", dsn)
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Errorf("Error closing database connection: %v", err)
		}
	}(db)

	songsWithMissingProperties, err := getSongsWithMissingProperties(db)
	if err != nil {
		log.Fatalf("Error retrieving songs with missing properties: %v", err)
	}

	batchSize := 50
	for i := 0; i < len(songsWithMissingProperties); i += batchSize {
		log.Infof("Processing batch %d to %d", i, i+batchSize)
		end := i + batchSize
		if end > len(songsWithMissingProperties) {
			end = len(songsWithMissingProperties)
		}

		batch := songsWithMissingProperties[i:end]
		var spotifyIds []spotify2.ID
		for _, song := range batch {
			spotifyIds = append(spotifyIds, spotify2.ID(song.SpotifyId))
		}

		tracks, err := spotifyClient.Client.GetTracks(context.Background(), spotifyIds)
		if err != nil {
			log.Fatalf("Error retrieving tracks from Spotify: %v", err)
		}

		for _, track := range tracks {
			if track == nil {
				log.Warnf("Track is nil, skipping")
				continue
			}
			// update the song properties in the database
			err := updateSongProperties(db, track.ID, int(track.Duration), int(track.Popularity), track.Album.Name)
			if err != nil {
				log.Errorf("Error updating song properties for song ID %v: %v", track.ID, err)
			}
		}
	}

	log.Info("All songs processed")
}

type SongPartial struct {
	Id        int    `db:"id"`
	SpotifyId string `db:"spotify_id"`
}

func getSongsWithMissingProperties(db *sqlx.DB) ([]SongPartial, error) {
	// query the songs table, retrieving records with missing properties for duration, popularity, and album name
	query := `
		SELECT id, spotify_id
		FROM songs
		WHERE duration IS NULL OR popularity IS NULL OR album_name IS NULL
	`
	rows, err := db.Queryx(query)
	if err != nil {
		return nil, err
	}

	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {
			log.Errorf("Error closing rows: %v", err)
		}
	}(rows)

	var songs []SongPartial
	for rows.Next() {
		var song SongPartial
		err := rows.StructScan(&song)
		if err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}

	return songs, nil
}

func updateSongProperties(db *sqlx.DB, spotifyTrackId spotify2.ID, duration int, popularity int, albumName string) error {
	// update the song record with the provided properties
	query := `
		UPDATE songs
		SET duration = $1, popularity = $2, album_name = $3
		WHERE spotify_id = $4
	`
	_, err := db.Exec(query, duration, popularity, albumName, spotifyTrackId)
	if err != nil {
		return err
	}
	return nil
}
