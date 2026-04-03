package main

import (
	"context"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/cron"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	_ = godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := cron.RunPurge(ctx); err != nil {
		log.Fatalf("Purge failed: %v", err)
	}
}
