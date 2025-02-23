package commands

import (
	"context"
	mongocommands "github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"time"
)

func SetMaxTrackAgePreference(durationInMilliseconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	err := mongocommands.SetPreference(ctx, types.MaxTrackAgeKey, durationInMilliseconds)
	cancel()
	if err != nil {
		return err
	}

	return nil
}

func GetMaxTrackAgePreference() (*types.Preference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	pref, err := mongocommands.GetPreference(ctx, types.MaxTrackAgeKey)
	cancel()
	if err != nil {
		return nil, err
	}

	return pref, nil
}
