package commands

import (
	"context"
	"github.com/vaultbotx/vaultbot-lite/internal/database/mongo"
	"github.com/vaultbotx/vaultbot-lite/internal/types"
	"go.mongodb.org/mongo-driver/bson"
)

func GetAllPreferences(ctx context.Context) (map[types.PreferenceKey]types.Preference, error) {
	instance, err := mongo.GetMongoClient(ctx)
	if err != nil {
		return nil, err
	}
	defer instance.Disconnect(ctx)

	collection := instance.Database(mongo.DatabaseName).Collection(mongo.PreferencesCollection)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	preferences := make(map[types.PreferenceKey]types.Preference)
	for cursor.Next(ctx) {
		var preference types.Preference
		err2 := cursor.Decode(&preference)
		if err2 != nil {
			return nil, err2
		}
		preferences[preference.Key] = preference
	}

	if err2 := cursor.Err(); err2 != nil {
		return nil, err2
	}

	return preferences, nil
}
