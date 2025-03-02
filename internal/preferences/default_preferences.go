package preferences

import (
	"context"
	log "github.com/sirupsen/logrus"
	mg "github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

func CheckDefaultPreferences(ctx context.Context) error {
	preferences, err := GetAllPreferences(ctx)
	if err != nil {
		return err
	}

	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return err
	}
	defer func(instance *mongo.Client, ctx context.Context) {
		err := instance.Disconnect(ctx)
		if err != nil {
			log.Errorf("Error disconnecting from MongoDB: %s", err)
			return
		}
	}(instance, ctx)

	for _, preferenceKey := range AllPreferences {
		if _, ok := preferences[preferenceKey]; !ok {
			log.Info("Preference %s does not exist, creating with default value", preferenceKey)
			err = SetPreference(instance, preferenceKey, preferenceKey.DefaultValue(), ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
