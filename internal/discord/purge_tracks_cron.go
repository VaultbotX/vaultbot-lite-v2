package discord

import (
	"context"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	internalcommands "github.com/vaultbotx/vaultbot-lite/internal/commands"
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
	pref, err := internalcommands.GetPurgeFrequencyPreference()
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
			pref, err2 := internalcommands.GetPurgeFrequencyPreference()
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

func purgeTracks() {
	log.Info("Purging tracks")
	newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	err := internalcommands.PurgeTracks(newCtx)
	if err != nil {
		log.Fatalf("Failed to purge tracks: %v", err)
	}
	cancel()
}
