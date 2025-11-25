package cron

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/tracks"
)

var (
	purgeJob      *gocron.Job
	purgeDuration time.Duration
)

func RunPurge(scheduler *gocron.Scheduler) {
	pref, err := getPurgeFrequencyPreference()
	if err != nil {
		log.Fatal(err)
	}

	num, err := pref.IntValue()
	if err != nil {
		log.Fatalf("Failed to convert preference value to int: %v", err)
	}

	purgeDuration = time.Duration(num) * time.Millisecond
	log.Infof("Scheduling purge tracks every %v", purgeDuration)
	purgeJob, err = scheduler.Every(purgeDuration).Do(purgeTracks)
	if err != nil {
		log.Fatalf("Failed to schedule purge tracks: %v", err)
	}

	frequencyChange := make(chan time.Duration)

	go func() {
		for {
			log.Debug("Checking for purge frequency changes")
			pref, err := getPurgeFrequencyPreference()
			if err != nil {
				log.Fatal(err)
			}

			num, err := pref.IntValue()
			if err != nil {
				log.Fatalf("Failed to convert preference value to int: %v", err)
			}

			newDuration := time.Duration(num) * time.Millisecond
			if newDuration != purgeDuration {
				frequencyChange <- newDuration
			}
			purgeDuration = newDuration

			time.Sleep(5 * time.Minute)
		}
	}()

	go func() {
		for {
			select {
			case newDuration := <-frequencyChange:
				log.Infof("Updating purge frequency to %v", purgeDuration)
				scheduler.Remove(purgeJob)
				purgeJob, err = scheduler.Every(newDuration).Do(purgeTracks)
				if err != nil {
					log.Fatalf("Failed to schedule purge tracks: %v", err)
				}
			}

			time.Sleep(5 * time.Minute)
		}
	}()
}

func getPurgeFrequencyPreference() (*domain.Preference, error) {
	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		return nil, err
	}

	preferenceService := domain.NewPreferenceService(persistence.NewPostgresPreferenceRepository(pgConn))
	return preferenceService.Repo.Get(context.Background(), domain.PurgeFrequencyKey)
}

func purgeTracks() {
	log.Info("Purging tracks")
	newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		log.Error(err)
		cancel()
		return
	}

	preferenceService := domain.NewPreferenceService(persistence.NewPostgresPreferenceRepository(pgConn))

	spClient, err := spotify.NewSpotifyClient(newCtx)
	if err != nil {
		log.Error(err)
		cancel()
		return
	}
	spPlaylistService := domain.NewSpotifyPlaylistService(&sp.SpotifyPlaylistRepo{
		Client: spClient,
	})

	now := time.Now().UTC()
	_, err = tracks.PurgeTracks(newCtx, now, preferenceService, spPlaylistService)
	if err != nil {
		log.Fatalf("Failed to purge tracks: %v", err)
	}
	cancel()
}
