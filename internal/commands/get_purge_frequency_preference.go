package commands

import (
	"context"
	mongocommands "github.com/vaultbotx/vaultbot-lite/internal/database/mongo/commands"
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
