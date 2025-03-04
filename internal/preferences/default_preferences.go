package preferences

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	mg "github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

func CheckDefaultPreferences(ctx context.Context) error {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		log.Errorf("Error getting MongoDB client: %s", err)
		return err
	}
	defer func(instance *mongo.Client, ctx context.Context) {
		err := instance.Disconnect(ctx)
		if err != nil {
			log.Errorf("Error disconnecting from MongoDB: %v", err)
			return
		}
	}(instance, ctx)
	preferenceService := domain.NewPreferenceService(persistence.PreferenceRepo{
		Client: instance,
	})

	preferences, err := preferenceService.Repo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, preferenceKey := range domain.AllPreferences {
		if _, ok := preferences[preferenceKey]; !ok {
			log.Info("Preference %s does not exist, creating with default value", preferenceKey)
			err = preferenceService.Repo.Set(ctx, preferenceKey, preferenceKey.DefaultValue())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
