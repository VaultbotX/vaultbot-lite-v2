package preferences

import (
	"context"
	log "github.com/sirupsen/logrus"
	mg "github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func SetMaxSongDurationPreference(durationInMilliseconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)

	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		cancel()
		return err
	}
	defer func(instance *mongo.Client, ctx context.Context) {
		err := instance.Disconnect(ctx)
		if err != nil {
			log.Errorf("Error disconnecting from MongoDB: %s", err)
			return
		}
	}(instance, ctx)

	err = SetPreference(instance, MaxDurationKey, durationInMilliseconds, ctx)
	cancel()
	if err != nil {
		return err
	}

	return nil
}

func GetMaxSongDurationPreference() (*Preference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	pref, err := GetPreference(ctx, MaxDurationKey)
	cancel()
	if err != nil {
		return nil, err
	}

	return pref, nil
}
