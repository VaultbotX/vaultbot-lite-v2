package cron

import (
	"os"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
)

var (
	genrePlaylistId string
)

func PopulateGenrePlaylist(scheduler *gocron.Scheduler) {
	playlistId, playlistIdExists := os.LookupEnv("GENRE_SPOTIFY_PLAYLIST_ID")
	if !playlistIdExists {
		log.Fatal("GENRE_SPOTIFY_PLAYLIST_ID must be set")
	}
	genrePlaylistId = playlistId

	_, err := scheduler.Every(1).Day().At("00:00").Do(populatePlaylist)
	if err != nil {
		log.Fatalf("Failed to schedule populate genre playlist job: %v", err)
	}
}

func populatePlaylist() {

}
