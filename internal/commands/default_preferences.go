package commands

import (
	"context"
	log "github.com/sirupsen/logrus"
	mongocommands "github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo/commands"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
)

func CheckDefaultPreferences(ctx context.Context) error {
	preferences, err := mongocommands.GetAllPreferences(ctx)
	if err != nil {
		return err
	}

	for _, preferenceKey := range types.AllPreferences {
		if _, ok := preferences[preferenceKey]; !ok {
			log.Info("Preference %s does not exist, creating with default value", preferenceKey)
			err = mongocommands.SetPreference(ctx, preferenceKey, preferenceKey.DefaultValue())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
