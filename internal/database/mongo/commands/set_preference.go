package commands

import (
	"context"
	"github.com/vaultbotx/vaultbot-lite/internal/database/mongo"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"go.mongodb.org/mongo-driver/bson"
)

func SetPreference(ctx context.Context, key types.PreferenceKey, value interface{}) error {
	instance, err := mongo.GetMongoClient(ctx)
	if err != nil {
		return err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mongo.DatabaseName).Collection(mongo.PreferencesCollection)

	filter := bson.M{"key": key}
	update := bson.M{"$set": bson.M{"value": value}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}
