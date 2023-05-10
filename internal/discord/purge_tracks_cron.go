package discord

import (
	"context"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	internalcommands "github.com/vaultbotx/vaultbot-lite/internal/commands"
	"time"
)

var job *gocron.Job

func RunPurge() {
	scheduler := gocron.NewScheduler(time.UTC)
	pref, err := internalcommands.GetPurgeFrequencyPreference()
	duration := time.Duration(pref.Value.(int))
	job, err = scheduler.Every(duration).Milliseconds().Do(purgeTracks)
	if err != nil {
		log.Fatalf("Failed to schedule purge tracks: %v", err)
	}

	frequencyChange := make(chan time.Duration)

	go func() {
		for {
			pref, err2 := internalcommands.GetPurgeFrequencyPreference()
			if err2 != nil {
				log.Fatal(err)
			}

			duration := time.Duration(pref.Value.(int))
			frequencyChange <- duration

			time.Sleep(5 * time.Minute)
		}
	}()

	scheduler.StartAsync()

	for {
		select {
		case duration := <-frequencyChange:
			log.Infof("Updating purge frequency to %v", duration)
			scheduler.Remove(job)
			job, err = scheduler.Every(duration).Milliseconds().Do(purgeTracks)
			if err != nil {
				log.Fatalf("Failed to schedule purge tracks: %v", err)
			}
		}

		time.Sleep(5 * time.Minute)
	}
}

func purgeTracks() func() {
	return func() {
		log.Info("Purging tracks")
		newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		err := internalcommands.PurgeTracks(newCtx)
		if err != nil {
			log.Fatalf("Failed to purge tracks: %v", err)
		}
		cancel()
	}
}
