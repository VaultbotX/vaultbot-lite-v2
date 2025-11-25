package cron

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
)

func PopulateGenrePlaylist(scheduler *gocron.Scheduler) {
	_, err := scheduler.Every(1).Day().At("00:00").Do(populatePlaylist)
	if err != nil {
		log.Fatalf("Failed to schedule populate genre playlist job: %v", err)
	}
}

func populatePlaylist() {
	newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	spClient, err := spotify.NewSpotifyClient(newCtx)
	if err != nil {
		log.Error(err)
		cancel()
		return
	}
	playlistService := domain.NewSpotifyPlaylistService(&sp.SpotifyPlaylistRepo{
		Client: spClient,
	})

	playlistItems, err := playlistService.Repo.GetPlaylistTracks(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	panic("not implemented")

	cancel()
}
