package discord

import (
	"context"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	internalcommands "github.com/vaultbotx/vaultbot-lite/internal/commands"
	"time"
)

func RunPurge() {
	s := gocron.NewScheduler(time.UTC)
	_, err := s.Every(12).Hours().Do(func() {
		newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		err := internalcommands.PurgeTracks(newCtx)
		if err != nil {
			log.Fatalf("Failed to purge tracks: %v", err)
		}
		cancel()
	})
	if err != nil {
		log.Fatalf("Failed to schedule purge tracks: %v", err)
	}

	s.StartAsync()
}
