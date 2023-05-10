package commands

import (
	"context"
	mongocommands "github.com/vaultbotx/vaultbot-lite/internal/database/mongo/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
)

func CheckDefaultPreferences(ctx context.Context) error {
	preferences, err := mongocommands.GetAllPreferences(ctx)
	if err != nil {
		return err
	}

	for _, preferenceKey := range types.AllPreferences {
		// If the preference doesn't exist, create it
		if _, ok := preferences[preferenceKey]; !ok {
			err = mongocommands.SetPreference(ctx, preferenceKey, preferenceKey.DefaultValue())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
