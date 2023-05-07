package commands

import (
	"context"
	mg "github.com/vaultbotx/vaultbot-lite/internal/database/mongo"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetPreference(ctx context.Context, key types.PreferenceKey) (interface{}, error) {
	instance, err := mg.GetMongoClient(ctx)
	if err != nil {
		return nil, err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mg.DatabaseName).Collection(mg.PreferencesCollection)

	filter := bson.M{"key": key}
	var preference map[string]interface{}
	err = collection.FindOne(ctx, filter).Decode(&preference)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return preference["value"], nil
}
