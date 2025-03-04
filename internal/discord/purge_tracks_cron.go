package discord

import (
	"context"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	mg "github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/tracks"
	"go.mongodb.org/mongo-driver/mongo"
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
	duration = time.Duration(pref.Value.(int32)) * time.Millisecond
	log.Infof("Scheduling purge tracks every %v", duration)
	job, err = scheduler.Every(duration).Do(purgeTracks)
	if err != nil {
		log.Fatalf("Failed to schedule purge tracks: %v", err)
	}

	frequencyChange := make(chan time.Duration)

	go func() {
		for {
			log.Debug("Checking for purge frequency changes")
			pref, err2 := getPurgeFrequencyPreference()
			if err2 != nil {
				log.Fatal(err)
			}

			newDuration := time.Duration(pref.Value.(int32)) * time.Millisecond
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
	instance, err := mg.GetMongoClient(context.Background())
	if err != nil {
		log.Errorf("Error getting MongoDB client: %s", err)
		return nil, err
	}
	defer func(instance *mongo.Client) {
		err := instance.Disconnect(context.Background())
		if err != nil {
			log.Errorf("Error disconnecting from MongoDB: %v", err)
			return
		}
	}(instance)

	preferenceService := domain.NewPreferenceService(persistence.PreferenceRepo{
		Client: instance,
	})

	return preferenceService.Repo.Get(context.Background(), domain.PurgeFrequencyKey)
}

func purgeTracks() {
	log.Info("Purging tracks")
	newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	instance, err := mg.GetMongoClient(newCtx)
	if err != nil {
		cancel()
		log.Errorf("Error getting MongoDB client: %s", err)
		return
	}
	defer func(instance *mongo.Client, ctx context.Context) {
		err := instance.Disconnect(ctx)
		if err != nil {
			log.Errorf("Error disconnecting from MongoDB: %v", err)
			return
		}
	}(instance, newCtx)

	preferenceService := domain.NewPreferenceService(persistence.PreferenceRepo{
		Client: instance,
	})

	spClient, err := spotify.NewSpotifyClient(newCtx)
	if err != nil {
		log.Error(err)
		cancel()
		return
	}
	spPlaylistService := domain.NewSpotifyPlaylistService(&sp.SpotifyPlaylistRepo{
		Client: spClient,
	})

	err = tracks.PurgeTracks(newCtx, preferenceService, spPlaylistService)
	if err != nil {
		log.Fatalf("Failed to purge tracks: %v", err)
	}
	cancel()
}
