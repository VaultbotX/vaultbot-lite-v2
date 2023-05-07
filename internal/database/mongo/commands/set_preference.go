package commands

import (
	"context"
	"github.com/vaultbotx/vaultbot-lite/internal/database/mongo"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	upsert := true
	opts := options.UpdateOptions{Upsert: &upsert}
	_, err = collection.UpdateOne(context.Background(), filter, update, &opts)
	if err != nil {
		return err
	}

	return nil
}
