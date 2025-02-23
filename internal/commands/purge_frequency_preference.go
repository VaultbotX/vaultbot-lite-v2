package commands

import (
	"context"
	mongocommands "github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"time"
)

func GetPurgeFrequencyPreference() (*types.Preference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	pref, err := mongocommands.GetPreference(ctx, types.PurgeFrequencyKey)
	cancel()
	if err != nil {
		return nil, err
	}

	return pref, nil
}

func SetPurgeFrequencyPreference(durationInMilliseconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	err := mongocommands.SetPreference(ctx, types.PurgeFrequencyKey, durationInMilliseconds)
	cancel()
	if err != nil {
		return err
	}

	return nil
}
