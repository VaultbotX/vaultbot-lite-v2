package commands

import (
	"context"
	"github.com/vaultbotx/vaultbot-lite/internal/database/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func GetAllPreferences(ctx context.Context) (map[string]interface{}, error) {
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

	var preferences map[string]interface{}
	for cursor.Next(ctx) {
		err2 := cursor.Decode(&preferences)
		if err2 != nil {
			return nil, err2
		}
	}

	if err2 := cursor.Err(); err2 != nil {
		return nil, err2
	}

	return preferences, nil
}
