package discord

import (
	"context"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/tracks"
	"time"
)

var (
	job      *gocron.Job
	duration time.Duration
)

// Deprecated: remove gocron in favor of https://github.com/hibiken/asynq
// so that we can have a more reliable task scheduler backed by Redis
// that also supports general event scheduling
func RunPurge() {
	scheduler := gocron.NewScheduler(time.UTC)
	pref, err := getPurgeFrequencyPreference()
	if err != nil {
		log.Fatal(err)
	}

	num, err := pref.IntValue()
	if err != nil {
		log.Fatalf("Failed to convert preference value to int: %v", err)
	}

	duration = time.Duration(num) * time.Millisecond
	log.Infof("Scheduling purge tracks every %v", duration)
	job, err = scheduler.Every(duration).Do(purgeTracks)
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
			if newDuration != duration {
				frequencyChange <- newDuration
			}
			duration = newDuration

			time.Sleep(5 * time.Minute)
		}
	}()

	scheduler.StartAsync()

	go func() {
		for {
			select {
			case newDuration := <-frequencyChange:
				log.Infof("Updating purge frequency to %v", duration)
				scheduler.Remove(job)
				job, err = scheduler.Every(newDuration).Do(purgeTracks)
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

	_, err = tracks.PurgeTracks(newCtx, preferenceService, spPlaylistService)
	if err != nil {
		log.Fatalf("Failed to purge tracks: %v", err)
	}
	cancel()
}
