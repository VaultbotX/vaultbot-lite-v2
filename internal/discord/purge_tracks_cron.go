package discord

import (
	"context"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	internalcommands "github.com/vaultbotx/vaultbot-lite/internal/commands"
	"time"
)

func RunPurge(ctx context.Context) {
	s := gocron.NewScheduler(time.UTC)
	_, err := s.Every(12).Hours().Do(func() {
		err := internalcommands.PurgeTracks(ctx)
		if err != nil {
			log.Fatalf("Failed to purge tracks: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("Failed to schedule purge tracks: %v", err)
	}

	s.StartAsync()
}
