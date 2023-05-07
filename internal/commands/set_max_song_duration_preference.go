package commands

import (
	"context"
	mongocommands "github.com/vaultbotx/vaultbot-lite/internal/database/mongo/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"time"
)

func SetMaxSongDurationPreference(durationInMilliseconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	err := mongocommands.SetPreference(ctx, types.MaxDurationKey, durationInMilliseconds)
	cancel()
	if err != nil {
		return err
	}

	return nil
}
